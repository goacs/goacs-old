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
	"net/url"
	"time"
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
	task.ParameterValues = []types.ParameterValueStruct{
		{
			Name: param,
		},
	}

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

	parsedUrl, err := url.Parse(acsRequest.CPE.ConnectionRequestUrl)
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(parsedUrl, prepareCookies(acsRequest.Session))

	client := http.Client{
		Timeout: time.Second * 5,
		Jar:     jar,
	}

	request.HTTPClient = &client

	response, err := request.Execute()

	if err != nil {
		log.Println("acs req error", err)
		return err
	}
	defer response.Body.Close()
	acsRequest.Response = response

	return nil

}
