package http

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	dac "github.com/xinsnake/go-http-digest-auth-client"
	"goacs/acs"
	acsxml "goacs/acs/types"
	"goacs/models/cpe"
	"goacs/repository"
	"log"
	"net/http"
	"time"
)

type ACSRequest struct {
	DBConnection *sqlx.DB
	CPE          *cpe.CPE
	Body         string
	Response     *http.Response
	Session      *acs.ACSSession
}

func NewACSRequest(cpe *cpe.CPE) *ACSRequest {
	request := &ACSRequest{
		DBConnection: repository.GetConnection(),
		CPE:          cpe,
		Body:         "",
	}

	request.Session = acs.CreateEmptySession(cpe.SerialNumber)

	return request
}

func (ACSRequest *ACSRequest) AddObject(param string) {
	envelope := acsxml.NewEnvelope()
	reqBody := envelope.AddObjectRequest(param, "")
	log.Println(reqBody)

	ACSRequest.Body = reqBody
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

	client := http.Client{
		Timeout: time.Second * 5,
	}

	request.Header.Add("Set-Cookie", fmt.Sprintf("%s=%s; ", "sessionId", acsRequest.Session.Id))
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
