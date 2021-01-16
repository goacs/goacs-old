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

	session := acs.GetSessionFromRequest(request)

	if session == nil {
		session = acs.CreateEmptySession(acs.GenerateSessionId())
	}

	if err != io.EOF && err != nil {
		return
	}

	reqType, envelope := parseBody(buffer)

	var reqRes = acshttp.CPERequest{
		Request:      request,
		Response:     w,
		DBConnection: repository.GetConnection(),
		Session:      session,
		Envelope:     &envelope,
		Body:         buffer,
		ReqType:      reqType,
	}

	if reqRes.Session != nil {
		acs.AddCookieToResponseWriter(reqRes.Session, reqRes.Response)
	}

	switch reqType {
	case acsxml.InformReq:
		informDecision := methods.InformDecision{&reqRes}
		informDecision.CpeInformRequestParser()

	case acsxml.EMPTY:
		log.Println("EMPTY RESPONSE")
		if len(session.Tasks) == 0 {
			acs.DeleteSession(session.Id)
		}

	case acsxml.GPNResp:
		parameterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		parameterDecisions.CpeParameterNamesResponseParser()

	case acsxml.GPVResp:
		parameterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		parameterDecisions.GetParameterValuesResponseParser()

	case acsxml.AddObjResp:
		log.Println("AddObjResp")
		log.Println(string(reqRes.Body))
		paramaterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		paramaterDecisions.AddObjectResponseParser()

	case acsxml.FaultResp:
		var faultresponse acsxml.Fault
		_ = xml.Unmarshal(buffer, &faultresponse)
		reqRes.Session.CPE.Fault = faultresponse
		faultDecision := methods.FaultDecision{ReqRes: &reqRes}
		faultDecision.ResponseDecision()
		if len(session.Tasks) == 0 {
			acs.DeleteSession(session.Id)
		}

	default:
		fmt.Println("UNSUPPORTED REQTYPE ", reqType)
	}

	if session.Provision == true {
		parameterDecisions := methods.ParameterDecisions{ReqRes: &reqRes}
		parameterDecisions.PrepareParametersToSend()
	}

	ProcessTasks(&reqRes, reqType)

}

func ProcessTasks(reqRes *acshttp.CPERequest, event string) {
	tasksRepository := mysql.NewTasksRepository(reqRes.DBConnection)
	cpeDatabaseTasks := tasksRepository.GetTasksForCPE(reqRes.Session.CPE.UUID)

	if reqRes.Session.IsNewInACS == true && event == acsxml.GPVResp {
		newDeviceTask := tasksRepository.GetGlobalTask("new")
		reqRes.Session.AddTask(newDeviceTask)
	}

	if len(cpeDatabaseTasks) > 0 {
		filteredTasks := tasks.FilterTasksByEvent(event, cpeDatabaseTasks)
		for _, cpeTask := range filteredTasks {
			reqRes.Session.AddTask(cpeTask)
		}
	}

	if len(reqRes.Session.Tasks) > 0 {
		for _, task := range reqRes.Session.Tasks {
			//log.Println("TASKS", reqRes.Session.Tasks)
			reqRes.Session.Tasks = reqRes.Session.Tasks[1:]
			waitForResponse := ProcessTask(task, reqRes)
			if waitForResponse == false {
				ProcessTasks(reqRes, event)
			} else {
				return
			}
		}
	}

}

func ProcessTask(task tasks.Task, reqRes *acshttp.CPERequest) bool {
	log.Println("Processing task: ", task.Task)
	if task.Task == tasks.RunScript {
		scriptEngine := scripts.NewScriptEngine(reqRes)
		_, err := scriptEngine.Execute(task.Script)
		log.Println(err)
		parameterDecisions := methods.ParameterDecisions{ReqRes: reqRes}
		parameterDecisions.PrepareParametersToSend()
		//log.Println("TASK QUEUE", reqRes.Session.Tasks)

		return false
	} else if task.Task == acsxml.InformResp {
		informMethods := methods.InformDecision{ReqRes: reqRes}
		body := informMethods.CpeInformResponse()
		reqRes.SendResponse(body)
	} else if task.Task == acsxml.GPNReq {
		parameterMethods := methods.ParameterDecisions{ReqRes: reqRes}
		body := parameterMethods.ParameterNamesRequest(task.ParameterInfo[0].Name, false)
		reqRes.SendResponse(body)
	} else if task.Task == acsxml.GPVReq {
		parameterMethods := methods.ParameterDecisions{ReqRes: reqRes}
		body := parameterMethods.GetParameterValuesRequest([]acsxml.ParameterInfo{
			{
				Name:     reqRes.Session.CPE.Root + ".",
				Writable: "0",
			},
		})
		reqRes.SendResponse(body)
	} else if task.Task == acsxml.SPVReq {
		body := reqRes.Envelope.SetParameterValues(reqRes.Session.CPE.PopParametersQueue())
		reqRes.SendResponse(body)
	}

	if task.Id != 0 && task.Infinite == false {
		tasksRepository := mysql.NewTasksRepository(reqRes.DBConnection)
		tasksRepository.DoneTask(task.Id)
	}

	return true
}

func parseBody(buffer []byte) (string, acsxml.Envelope) {
	//fmt.Println(string(buffer))
	var envelope acsxml.Envelope
	err := xml.Unmarshal(buffer, &envelope)

	var requestType string = acsxml.EMPTY

	if err == nil {
		switch envelope.Type() {
		case "inform":
			requestType = acsxml.InformReq
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
