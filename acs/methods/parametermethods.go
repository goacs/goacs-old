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
	"strings"
)

const MAX_GPN_REQUESTS = 50
const MAX_CHUNK_SIZE = 1

type ParameterDecisions struct {
	ReqRes *http.CPERequest
}

func (pd *ParameterDecisions) ParameterNamesRequest(path string, nextlevel bool) string {
	pd.ReqRes.Session.PrevReqType = acsxml.GPNReq
	return pd.ReqRes.Envelope.GPNRequest(path, nextlevel)
}

func (pd *ParameterDecisions) CpeParameterNamesResponseParser() {
	//acsxml.PrintParamsInfo(pd.ReqRes.Session.CPE.ParametersInfo, pd.ReqRes.Session.CPE.SerialNumber)

	var gpnr acsxml.GetParameterNamesResponse
	log.Println("CpeParameterNamesResponseParser")

	_ = xml.Unmarshal(pd.ReqRes.Body, &gpnr)
	pd.ReqRes.Session.CPE.AddParametersInfo(gpnr.ParameterList)
	pd.ReqRes.Session.GPNCount--

	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_ = cpeRepository.BulkInsertOrUpdateParameters(&pd.ReqRes.Session.CPE, pd.ReqRes.Session.CPE.GetObjectNamesToParameters())

	nextLevelParams := pd.GetNextLevelParams(pd.ReqRes.Session.CPE.ParametersInfo)

	if pd.ReqRes.Session.GPNCount < MAX_GPN_REQUESTS && len(nextLevelParams) > 0 {
		for _, nextLevelParam := range nextLevelParams {
			if pd.ReqRes.Session.GPNCount > MAX_GPN_REQUESTS {
				continue
			}

			parameterInfo := acsxml.ParameterInfo{
				Name:     nextLevelParam.Name,
				Writable: nextLevelParam.Writable,
				Done:     false,
			}

			task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
			task.Task = acsxml.GPNReq
			task.ParameterInfo = append(task.ParameterInfo, parameterInfo)
			pd.ReqRes.Session.AddTask(task)
			pd.ReqRes.Session.AddParameterNamesToQueryValues(parameterInfo)
			pd.ReqRes.Session.GPNCount++
			log.Println("CURRENT GPN COUNT", pd.ReqRes.Session.GPNCount)
			log.Println("added task", task)
		}

		return //if we have nextLevelParams, then prevent GPVReq add task
	}
	log.Println("Current GPN tasks", pd.ReqRes.Session.Tasks)
	if (pd.ReqRes.Session.IsNewInACS || pd.ReqRes.Session.IsBoot) && len(pd.ReqRes.Session.Tasks) == 0 {
		log.Println("ADDING GPVREQ for these params", pd.ReqRes.Session.ParameterNamesToQueryValues)
		for _, parameterNames := range acsxml.ChunkParameterInfo(pd.ReqRes.Session.ParameterNamesToQueryValues, MAX_CHUNK_SIZE) {
			log.Println("adding gpvreq for new device")
			task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
			task.Task = acsxml.GPVReq
			task.ParameterInfo = parameterNames
			pd.ReqRes.Session.AddTask(task)
			pd.ReqRes.Session.GPVCount++
		}
	}

}

func (pd *ParameterDecisions) GetNextLevelParams(params []acsxml.ParameterInfo) []acsxml.ParameterInfo {
	var newParams []acsxml.ParameterInfo
	for idx, param := range params {
		if needsToQueryParam(param) {
			log.Println("GetNExtLevelParamName", pd.ReqRes.Session.CPE.ParametersInfo[idx].Name)
			log.Println("DONE", pd.ReqRes.Session.CPE.ParametersInfo[idx].Done)
			pd.ReqRes.Session.CPE.ParametersInfo[idx].Done = true
			newParams = append(newParams, param)
		}
	}
	return newParams
}

func needsToQueryParam(param acsxml.ParameterInfo) bool {
	if param.Name[len(param.Name)-1:] != "." {
		return false
	}
	chunks := strings.Split(param.Name, ".")
	return len(chunks) > 2 && len(chunks) <= 3 && param.Done == false
}

/*func (pd *ParameterDecisions) getParamsBeginningWith(path string) []acsxml.ParameterInfo {
	var params []acsxml.ParameterInfo
	for idx, param := range pd.ReqRes.Session.CPE.ParametersInfo {
		if param.Done == false && param.Writable == "0" && param.Name[len(param.Name)-1:] == "." {
			pd.ReqRes.Session.CPE.ParametersInfo[idx].Done = true
			params = append(params, param)
		}
	}
}*/

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

	pd.ReqRes.Session.GPVCount--

	//log.Println(pd.CPERequest.Session.CPE.ParameterValues)
	if pd.ReqRes.Session.IsNewInACS {
		_ = cpeRepository.BulkInsertOrUpdateParameters(&pd.ReqRes.Session.CPE, pd.ReqRes.Session.CPE.ParameterValues)
	} else if pd.ReqRes.Session.IsBoot {
		if pd.ReqRes.Session.HasTaskOfType(acsxml.GPVReq) == false {
			cpeObjectParameters := pd.ReqRes.Session.CPE.GetObjectNamesToParameters()
			dbObjectParameters := cpeRepository.GetCPEParametersWithFlag(&pd.ReqRes.Session.CPE, "A")

			addparams, delparams := cpe.CompareObjectParameters(cpeObjectParameters, dbObjectParameters)

			if len(delparams) > 0 {
				for _, param := range delparams {
					task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
					task.Task = acsxml.DelObjReq
					task.ParameterValues = []acsxml.ParameterValueStruct{param}
					log.Println("ADDING DELOBJECT TASK", task)
					//pd.ReqRes.Session.AddTask(task)
				}

			}

			if len(addparams) > 0 {
				for _, param := range addparams {
					newName, err := acsxml.ObjectParamToInstance(param.Name)

					if err != nil {
						log.Println("addparams error", err)
						continue
					}
					param.Name = newName
					task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
					task.Task = acsxml.AddObjReq
					task.ParameterValues = []acsxml.ParameterValueStruct{param}
					log.Println("ADDING ADDOBJECT TASK", task)
					//pd.ReqRes.Session.AddTask(task)
				}

			}

		}
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
			pd.ReqRes.Session.ParametersToAdd = append(pd.ReqRes.Session.ParametersToAdd, diffParameters...)
		}
	}

	//TODO: Check why some parameters are writeable, but cpe returns fault on it
	if len(pd.ReqRes.Session.ParametersToAdd) > 0 {
		log.Println("ADD SPVREQ TASK")
		task := tasks.NewCPETask(pd.ReqRes.Session.CPE.UUID)
		task.Task = acsxml.SPVReq
		pd.ReqRes.Session.AddTask(task)
		pd.ReqRes.Session.SPVCount++
	}
}

func (pd *ParameterDecisions) AddObjectResponseParser() acsxml.AddObjectResponseStruct {
	var addObject acsxml.AddObjectResponseStruct
	_ = xml.Unmarshal(pd.ReqRes.Body, &addObject)

	return addObject
}
