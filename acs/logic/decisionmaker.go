package logic

import (
	"encoding/xml"
	"fmt"
	"goacs/acs"
	acshttp "goacs/acs/http"
	"goacs/acs/methods"
	"goacs/acs/scripts"
	acsxml "goacs/acs/types"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func CPERequestDecision(request *http.Request, w http.ResponseWriter) {
	buffer, err := ioutil.ReadAll(request.Body)
	//session := acs.CreateSession(request)

	//w = acs.AddCookieToResponseWriter(session, w)

	if err != io.EOF && err != nil {
		return
	}

	reqType, envelope := parseEnvelope(buffer)

	var reqRes = acshttp.CPERequest{
		Request:      request,
		Response:     w,
		DBConnection: repository.GetConnection(),
		Session:      acs.GetSessionFromRequest(request),
		Envelope:     &envelope,
		Body:         buffer,
	}

	if reqRes.Session != nil {
		acs.AddCookieToResponseWriter(reqRes.Session, reqRes.Response)
	}

	switch reqType {
	case acsxml.INFORM:
		informDecision := methods.InformDecision{&reqRes}
		informDecision.CpeInformRequestParser()
		informDecision.CpeInformResponse()

	case acsxml.EMPTY:
		log.Println("EMPTY RESPONSE")
		if reqRes.Session.PrevReqType == acsxml.INFORM {
			if reqRes.Session.ReadAllParameters == true {
				reqRes.Session.NextJob = acs.JOB_GETPARAMETERNAMES
			}
		}

	case acsxml.GPNResp:
		parameterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		parameterDecisions.CpeParameterNamesResponseParser()
		reqRes.Session.NextJob = acs.JOB_GETPARAMETERVALUES
		log.Println("GPNResponse next job", reqRes.Session.NextJob)

	case acsxml.GPVResp:
		parameterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		parameterDecisions.GetParameterValuesResponseParser()

	case acsxml.SPVResp:
		paramaterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		paramaterDecisions.SetParameterValuesRequest()

	case acsxml.AddObjResp:
		log.Println("AddObjResp")
		paramaterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		addObjectResponseStruct := paramaterDecisions.AddObjectResponseParser()

		if addObjectResponseStruct.Status == 0 {
			//TODO: make to get only parameters that was created by addObject
			reqRes.Session.NextJob = acs.JOB_GETPARAMETERNAMES
		}

	case acsxml.FaultResp:
		var faultresponse acsxml.Fault
		_ = xml.Unmarshal(buffer, &faultresponse)
		reqRes.Session.CPE.Fault = faultresponse
		faultDecision := methods.FaultDecision{ReqRes: &reqRes}
		faultDecision.ResponseDecision()

	default:
		fmt.Println("UNSUPPORTED REQTYPE ", reqType)
	}

	ProcessTasks(&reqRes, reqType)
	ProcessSessionJobs(&reqRes)

}

func ProcessSessionJobs(reqRes *acshttp.CPERequest) {
	log.Println("Processing next job", reqRes.Session.NextJob)
	switch reqRes.Session.NextJob {
	case acs.JOB_SENDPARAMETERS:
		parameterDecisions := methods.ParameterDecisions{ReqRes: reqRes}
		parameterDecisions.SetParameterValuesRequest()
		reqRes.Session.NextJob = acs.JOB_NONE

	case acs.JOB_GETPARAMETERNAMES:
		parameterDecisions := methods.ParameterDecisions{ReqRes: reqRes}
		parameterDecisions.ParameterNamesRequest(true)
		reqRes.Session.NextJob = acs.JOB_GETPARAMETERVALUES

	case acs.JOB_GETPARAMETERVALUES:
		log.Println("JOB_GETPARAMETERVALUES ROOT", reqRes.Session.CPE.Root)
		parameterDecision := methods.ParameterDecisions{ReqRes: reqRes}
		parameterDecision.GetParameterValuesRequest([]acsxml.ParameterInfo{
			{
				Name:     reqRes.Session.CPE.Root + ".",
				Writable: "0",
			},
		})
		reqRes.Session.NextJob = acs.JOB_NONE
	}
}

func ProcessTasks(reqRes *acshttp.CPERequest, event string) {
	tasksRepository := mysql.NewTasksRepository(reqRes.DBConnection)
	cpeTasks := tasksRepository.GetTasksForCPE(reqRes.Session.CPE.UUID)

	if reqRes.Session.IsNewInACS == true && event == acsxml.GPVResp {
		tasksForNewDevices := tasksRepository.GetGlobalTask("new")
		log.Println(tasksForNewDevices)
		cpeTasks = append(cpeTasks, tasksForNewDevices...)
	}

	if len(cpeTasks) > 0 {
		scriptEngine := scripts.NewScriptEngine(reqRes)
		filteredTasks := tasks.FilterTasksByEvent(event, cpeTasks)
		for _, cpeTask := range filteredTasks {
			switch cpeTask.Task {
			case tasks.RunScript:
				result, err := scriptEngine.Execute(cpeTask.Script)
				log.Println(err)
				log.Println(result)
			case tasks.SendParameters:
				scriptEngine.SendParameters()
			}

			if cpeTask.Infinite == false {
				tasksRepository.DoneTask(cpeTask.Id)
			}
		}
	}
}

func parseEnvelope(buffer []byte) (string, acsxml.Envelope) {
	//fmt.Println(string(buffer))
	var envelope acsxml.Envelope
	err := xml.Unmarshal(buffer, &envelope)

	var requestType string = acsxml.EMPTY

	if err == nil {
		switch envelope.Type() {
		case "inform":
			requestType = acsxml.INFORM
		case "getparameternamesresponse":
			requestType = acsxml.GPNResp
		case "getparametervaluesresponse":
			requestType = acsxml.GPVResp
		case "setparametervaluesresponse":
			requestType = acsxml.SPVResp
		case "addobjectresponse":
			requestType = acsxml.AddObjResp
		case "fault":
			requestType = acsxml.FaultResp
		default:
			fmt.Println("UNSUPPORTED envelope type " + envelope.Type())
			requestType = acsxml.UNKNOWN
		}
	}

	return requestType, envelope
}
