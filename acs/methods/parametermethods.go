package methods

import (
	"encoding/xml"
	"goacs/acs/http"
	acsxml "goacs/acs/types"
	"goacs/models/cpe"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
	"log"
)

type ParameterDecisions struct {
	ReqRes *http.CPERequest
}

func (pd *ParameterDecisions) ParameterNamesRequest(path string, nextlevel bool) string {
	pd.ReqRes.Session.PrevReqType = acsxml.GPNReq
	return pd.ReqRes.Envelope.GPNRequest(path, nextlevel)
}

func (pd *ParameterDecisions) CpeParameterNamesResponseParser() {
	acsxml.PrintParamsInfo(pd.ReqRes.Session.CPE.ParametersInfo, pd.ReqRes.Session.CPE.SerialNumber)

	var gpnr acsxml.GetParameterNamesResponse
	log.Println("CpeParameterNamesResponseParser")

	_ = xml.Unmarshal(pd.ReqRes.Body, &gpnr)
	pd.ReqRes.Session.CPE.AddParametersInfo(gpnr.ParameterList)

	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_ = cpeRepository.BulkInsertOrUpdateParameters(&pd.ReqRes.Session.CPE, pd.ReqRes.Session.CPE.GetAddObjectParameters())

	nextLevelParams := pd.getNextLevelParams()
	if len(nextLevelParams) > 0 {
		for _, nextLevelParam := range nextLevelParams {
			task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
			task.Task = acsxml.GPNReq
			task.ParameterInfo = append(task.ParameterInfo, acsxml.ParameterInfo{
				Name:     nextLevelParam.Name,
				Writable: nextLevelParam.Writable,
				Done:     false,
			})
			pd.ReqRes.Session.AddTask(task)
			log.Println("added task", task)
		}

		return //if we have nextLevelParams, then prevent GPVReq add task
	}

	//if pd.ReqRes.Session.IsNewInACS {
	//	log.Println("adding gpvreq for new device")
	//	task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
	//	task.Task = acsxml.GPVReq
	//	pd.ReqRes.Session.AddTask(task)
	//}

}

func (pd *ParameterDecisions) getNextLevelParams() []acsxml.ParameterInfo {
	var params []acsxml.ParameterInfo
	for idx, param := range pd.ReqRes.Session.CPE.ParametersInfo {
		if param.Done == false && param.Writable == "0" && param.Name[len(param.Name)-1:] == "." {
			pd.ReqRes.Session.CPE.ParametersInfo[idx].Done = true
			params = append(params, param)
		}
	}

	return params
}

func (pd *ParameterDecisions) GetParameterValuesRequest(parameters []acsxml.ParameterInfo) string {
	var request = pd.ReqRes.Envelope.GPVRequest(parameters)
	pd.ReqRes.Session.PrevReqType = acsxml.GPVReq
	return request
}

func (pd *ParameterDecisions) GetParameterValuesResponseParser() {
	var gpvr acsxml.GetParameterValuesResponse
	_ = xml.Unmarshal(pd.ReqRes.Body, &gpvr)
	log.Println("GetParameterValuesResponseParser")
	pd.ReqRes.Session.CPE.AddParameterValues(gpvr.ParameterList)
	pd.ReqRes.Session.FillCPESessionBaseInfo(gpvr.ParameterList)
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_, _, _ = cpeRepository.UpdateOrCreate(&pd.ReqRes.Session.CPE)

	//log.Println(pd.CPERequest.Session.CPE.ParameterValues)
	if pd.ReqRes.Session.IsNewInACS {
		_ = cpeRepository.BulkInsertOrUpdateParameters(&pd.ReqRes.Session.CPE, pd.ReqRes.Session.CPE.ParameterValues)
	}

}

func (pd *ParameterDecisions) PrepareParametersToSend() {
	pd.ReqRes.Session.PrevReqType = acsxml.SPVReq
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	templateRepository := mysql.NewTemplateRepository(repository.GetConnection())

	cpeDBParameters, err := cpeRepository.GetCPEParameters(&pd.ReqRes.Session.CPE)
	if err != nil {
		log.Println("Error GetParameterValuesResponseParser ", err.Error())
	}

	templateParameters := templateRepository.GetPrioritizedParametersForCPE(&pd.ReqRes.Session.CPE)
	cpeDBParameters = cpe.CombineTemplateParameters(cpeDBParameters, templateParameters)

	if len(cpeDBParameters) > 0 {
		//Get modified parameters
		//Check for AddObject instances
		diffParameters := pd.ReqRes.Session.CPE.GetChangedParametersToWrite(&cpeDBParameters)
		log.Println("DIFF PARAMS", diffParameters)
		if len(diffParameters) > 0 {
			pd.ReqRes.Session.CPE.ParametersQueue = append(pd.ReqRes.Session.CPE.ParametersQueue, diffParameters...)
		}
	}

	//TODO: Check why some parameters are writeable, but cpe returns fault on it
	if len(pd.ReqRes.Session.CPE.ParametersQueue) > 0 {
		log.Println("ADD SPVREQ TASK")
		task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
		task.Task = acsxml.SPVReq
		pd.ReqRes.Session.AddTask(task)
	}
}

func (pd *ParameterDecisions) AddObjectResponseParser() acsxml.AddObjectResponseStruct {
	var addObject acsxml.AddObjectResponseStruct
	_ = xml.Unmarshal(pd.ReqRes.Body, &addObject)

	return addObject
}
