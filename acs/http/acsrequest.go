package http

import (
	"github.com/jmoiron/sqlx"
	dac "github.com/xinsnake/go-http-digest-auth-client"
	"goacs/acs"
	"goacs/acs/types"
	"goacs/models/cpe"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
	"log"
	"net/http"
	"net/http/cookiejar"
)

type ACSRequest struct {
	DBConnection *sqlx.DB
	CPE          *cpe.CPE
	Body         string
	Response     *http.Response
	Session      *acs.ACSSession
	Jar          *cookiejar.Jar
}

func NewACSRequest(cpe *cpe.CPE) *ACSRequest {
	request := &ACSRequest{
		DBConnection: repository.GetConnection(),
		CPE:          cpe,
		Body:         "",
	}

	request.Session = acs.CreateEmptySession(acs.GenerateSessionId())

	return request
}

func prepareCookies(session *acs.ACSSession) []*http.Cookie {
	var cookies []*http.Cookie

	cookie := &http.Cookie{
		Name:  "sessionId",
		Value: session.Id,
	}

	cookies = append(cookies, cookie)

	return cookies
}

func (ACSRequest *ACSRequest) AddObject(param string) {
	taskrepository := mysql.NewTasksRepository(repository.GetConnection())
	task := tasks.NewCPETask(ACSRequest.CPE.UUID)

	task.Task = types.AddObjReq
	task.Event = types.InformReq
	task.AsAddObject(param)
	taskrepository.AddTask(task)
	err := ACSRequest.Send()

	if err != nil {
		log.Println("AddObject error", err.Error())
	}

}

func (ACSRequest *ACSRequest) GetParameterValues(path string) {
	ACSRequest.Session.NextJob = acs.JOB_GETPARAMETERNAMES
	err := ACSRequest.Send()

	if err != nil {
		log.Println("GetParameterValues error", err.Error())
	}
}

func (ACSRequest *ACSRequest) SetParameterValues() {
	ACSRequest.Session.NextJob = acs.JOB_SENDPARAMETERS
	err := ACSRequest.Send()

	if err != nil {
		log.Println("SetParameterValues error", err.Error())
	}
}

func (acsRequest *ACSRequest) Kick() {
	err := acsRequest.Send()

	if err != nil {
		log.Println("Kick error", err.Error())
	}

}

func (acsRequest *ACSRequest) Send() error {
	request := dac.NewRequest(acsRequest.CPE.ConnectionRequestUser, acsRequest.CPE.ConnectionRequestPassword, "GET", acsRequest.CPE.ConnectionRequestUrl, acsRequest.Body)

	log.Println(request)
	response, err2 := request.Execute()
	log.Println(response)

	if err2 != nil {
		log.Println("acs req error", err2)
		return err2
	}
	defer response.Body.Close()
	acsRequest.Response = response

	return nil

}
