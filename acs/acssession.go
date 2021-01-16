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
	Id                string
	IsNew             bool
	IsNewInACS        bool
	IsBoot            bool
	IsBootstrap       bool
	Provision         bool
	ReadAllParameters bool
	PrevReqType       string
	CreatedAt         time.Time
	CPE               cpe.CPE
	NextJob           int
	Tasks             []tasks.Task
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
	session.IsBoot = inform.IsBootEvent()
	session.IsBootstrap = inform.IsBootstrapEvent()
	session.FillCPESessionBaseInfo(inform.ParameterList)
}

func (session *ACSSession) FillCPESessionBaseInfo(parameters []types.ParameterValueStruct) {
	session.CPE.AddParameterValues(parameters)
	session.CPE.ConnectionRequestUrl, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.ConnectionRequestURL")
	session.CPE.ConnectionRequestUser, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.Username")
	session.CPE.ConnectionRequestPassword, _ = session.CPE.GetParameterValue(session.CPE.Root + ".ManagementServer.Password")
	session.CPE.HardwareVersion, _ = session.CPE.GetParameterValue(session.CPE.Root + ".DeviceInfo.HardwareVersion")
	session.CPE.SoftwareVersion, _ = session.CPE.GetParameterValue(session.CPE.Root + ".DeviceInfo.SoftwareVersion")
	ipAddrStr, _ := session.CPE.GetParameterValue(session.CPE.Root + "..WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress")
	_ = session.CPE.IpAddress.Scan(ipAddrStr)
}

func (session *ACSSession) AddTask(task tasks.Task) {
	session.Tasks = append(session.Tasks, task)
}
