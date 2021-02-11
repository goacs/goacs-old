package acs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	digest_auth_client "github.com/xinsnake/go-http-digest-auth-client"
	"goacs/acs/types"
	"goacs/models/cpe"
	"goacs/models/tasks"
	"log"
	"net/http"
	"sync"
	"time"
)

const SessionLifetime = 300
const SessionGoroutineTimeout = 10

const (
	JOB_NONE               = 0
	JOB_GETPARAMETERNAMES  = 1
	JOB_GETPARAMETERVALUES = 2
	JOB_SENDPARAMETERS     = 3
)

type ACSSession struct {
	Id                          string
	IsNew                       bool
	IsNewInACS                  bool
	IsBoot                      bool
	IsBootstrap                 bool
	Provision                   bool
	ReadAllParameters           bool
	CurrentState                string
	PrevState                   string
	CreatedAt                   time.Time
	CPE                         cpe.CPE
	NextJob                     int
	Tasks                       []tasks.Task
	GPNCount                    int //Count GPN requests to prevent many requests
	GPVCount                    int //Count GPV requests
	SPVCount                    int //Count GPV requests
	RunScriptCount              int
	RunnedScripts               int
	ParameterNamesToQueryValues []types.ParameterInfo
	ParametersToAdd             []types.ParameterValueStruct
	ParametersToDelete          []types.ParameterValueStruct
}

var lock = sync.RWMutex{}
var acsSessions map[string]*ACSSession

func GenerateSessionId() string {
	buff := make([]byte, 32)
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:32]
}

func StartSession() {
	fmt.Println("acsSessions init")
	acsSessions = make(map[string]*ACSSession)
	go removeOldSessions()
}

func GetSessionFromRequest(request *http.Request) *ACSSession {
	var session *ACSSession
	cookie, err := request.Cookie("sessionId")

	if err != nil {
		log.Println("cannot get cookie from request")
		return nil
	}
	session = GetSessionById(cookie.Value)
	if session != nil {
		session.IsNew = false
	}

	return session
}

func GetSessionById(sessionId string) *ACSSession {
	lock.RLock()
	session := acsSessions[sessionId]
	lock.RUnlock()

	if session != nil {
		log.Println("Session from memory ", sessionId)
	}

	return session
}

func AddCookieToResponseWriter(session *ACSSession, w http.ResponseWriter) http.ResponseWriter {
	cookie := http.Cookie{Name: "sessionId", Value: session.Id}
	http.SetCookie(w, &cookie)

	return w
}

func AddCookieToRequest(session *ACSSession, r *http.Request) {
	cookie := http.Cookie{Name: "sessionId", Value: session.Id}
	r.AddCookie(&cookie)
}

func AddCookieToDigestRequest(session *ACSSession, r *digest_auth_client.DigestRequest) {
	cookie := http.Cookie{Name: "sessionId", Value: session.Id}
	r.Header.Add("Set-Cookie", cookie.String())
}

func GetOrCreateSession(sessionId string) *ACSSession {
	var session *ACSSession
	session = GetSessionById(sessionId)

	if session == nil {
		session = CreateEmptySession(sessionId)
	}

	return session
}

func CreateEmptySession(sessionId string) *ACSSession {
	log.Println("creating new session", sessionId)
	session := ACSSession{Id: sessionId, IsNew: true, CreatedAt: time.Now()}
	lock.Lock()
	acsSessions[sessionId] = &session
	lock.Unlock()
	return acsSessions[sessionId]
}

func DeleteSession(sessionId string) {
	lock.Lock()
	delete(acsSessions, sessionId)
	lock.Unlock()
}

func removeOldSessions() {
	for {
		now := time.Now()
		for sessionId, session := range acsSessions {
			if now.Sub(session.CreatedAt).Minutes() > SessionLifetime {
				fmt.Println("DELETING OLD SESSION " + sessionId)
				DeleteSession(sessionId)
			}
		}
		time.Sleep(SessionGoroutineTimeout * time.Second)
	}
}

func (session *ACSSession) FillCPESessionFromInform(inform types.Inform) {
	session.CPE.SetRoot(cpe.DetermineDeviceTreeRootPath(inform.ParameterList))
	session.CPE.SerialNumber = inform.DeviceId.SerialNumber
	session.IsBoot = inform.IsBootEvent() || inform.IsBootstrapEvent()
	session.IsBootstrap = inform.IsBootstrapEvent()
	session.CPE.AddParameterValues(inform.ParameterList)
	session.FillCPESessionBaseInfo(inform.ParameterList)
}

func (session *ACSSession) FillCPESessionBaseInfo(parameters []types.ParameterValueStruct) {
	var value string
	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.ConnectionRequestURL")

	if value != "" {
		session.CPE.ConnectionRequestUrl = value
	}

	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.ConnectionRequestUsername")
	log.Println("ConnectionRequestUsername", value != "", value)
	if value != "" {
		log.Println("Changing username")
		session.CPE.ConnectionRequestUser = value
	}

	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.ConnectionRequestPassword")
	log.Println("ConnectionRequestPassword", value != "", value)

	if value != "" {
		session.CPE.ConnectionRequestPassword = value
	}

	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".DeviceInfo.HardwareVersion")
	if value != "" {
		session.CPE.HardwareVersion = value
	}

	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".DeviceInfo.SoftwareVersion")
	if value != "" {
		session.CPE.SoftwareVersion = value
	}

	value, _ = session.CPE.GetParameterValue(session.CPE.Root + ".WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress")
	if value != "" {
		_ = session.CPE.IpAddress.Scan(value)
	}
}

func (session *ACSSession) AddParameterNamesToQueryValues(param types.ParameterInfo) {
	for _, existParam := range session.ParameterNamesToQueryValues {
		if existParam.Name == param.Name {
			return
		}
	}

	session.ParameterNamesToQueryValues = append(session.ParameterNamesToQueryValues, param)
}

func (session *ACSSession) AddParameterToAdd(param types.ParameterValueStruct) {
	for _, existParam := range session.ParametersToAdd {
		if existParam.Name == param.Name {
			return
		}
	}

	session.ParametersToAdd = append(session.ParametersToAdd, param)
}

func (session *ACSSession) AddParameterToDelete(param types.ParameterValueStruct) {
	for _, existParam := range session.ParametersToDelete {
		if existParam.Name == param.Name {
			return
		}
	}

	session.ParametersToDelete = append(session.ParametersToDelete, param)
}

func (session *ACSSession) PopParametersToAdd() []types.ParameterValueStruct {

	defer func() {
		session.ParametersToAdd = []types.ParameterValueStruct{}
	}()

	return session.ParametersToAdd
}

func (session *ACSSession) AddTask(task tasks.Task) {
	if task.Task == tasks.RunScript {
		session.RunScriptCount++
	}

	session.Tasks = append(session.Tasks, task)
}

func (session *ACSSession) TaskExist(task tasks.Task) bool {
	if task.Id == 0 {
		return false
	}

	for _, sessionTask := range session.Tasks {
		if sessionTask.Id == task.Id {
			return true
		}
	}

	return false
}

func (session *ACSSession) HasTaskOfType(taskType string) bool {
	for _, sessionTask := range session.Tasks {
		if sessionTask.Task == taskType {
			return true
		}
	}

	return false
}
