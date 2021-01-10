package methods

import (
	"encoding/xml"
	"fmt"
	"goacs/acs/http"
	acsxml "goacs/acs/types"
	"goacs/models/cpe"
	"goacs/repository"
	"goacs/repository/mysql"
	"log"
)

type ParameterDecisions struct {
	ReqRes *http.CPERequest
}

func (pd *ParameterDecisions) ParameterNamesRequest(recursively bool) {
	pd.ReqRes.Session.PrevReqType = acsxml.GPNReq
	root := pd.ReqRes.Session.CPE.Root
	if recursively {
		root = root + "."
	}
	var request = pd.ReqRes.Envelope.GPNRequest(root)
	//fmt.Println(request)
	_, _ = fmt.Fprint(pd.ReqRes.Response, request)

}

func (pd *ParameterDecisions) CpeParameterNamesResponseParser() {
	var gpnr acsxml.GetParameterNamesResponse
	log.Println("CpeParameterNamesResponseParser")

	_ = xml.Unmarshal(pd.ReqRes.Body, &gpnr)
	pd.ReqRes.Session.CPE.AddParametersInfo(gpnr.ParameterList)

	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_ = cpeRepository.BulkInsertOrUpdateParameters(&pd.ReqRes.Session.CPE, pd.ReqRes.Session.CPE.GetAddObjectParameters())

}

func (pd *ParameterDecisions) GetParameterValuesRequest(parameters []acsxml.ParameterInfo) {
	var request = pd.ReqRes.Envelope.GPVRequest(parameters)
	_, _ = fmt.Fprint(pd.ReqRes.Response, request)
	pd.ReqRes.Session.PrevReqType = acsxml.GPVReq

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

func (pd *ParameterDecisions) SetParameterValuesRequest() {
	//parametersToWrite := pd.CPERequest.Session.CPE.GetParametersWithFlag("W")
	//log.Println("parametersToWrite")
	////log.Println(parametersToWrite)

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
		if len(diffParameters) > 0 {
			pd.ReqRes.Session.CPE.ParametersQueue = diffParameters
			//pd.ReqRes.Session.NextJob = acs.JOB_SENDPARAMETERS
		}
	}

	//TODO: Check why some parameters are writeable, but cpe returns fault on it
	if len(pd.ReqRes.Session.CPE.ParametersQueue) > 0 {
		log.Println("SPV")
		var response = pd.ReqRes.Envelope.SetParameterValues(pd.ReqRes.Session.CPE.PopParametersQueue())
		log.Println(response)
		_, _ = fmt.Fprint(pd.ReqRes.Response, response)
		pd.ReqRes.Session.PrevReqType = acsxml.SPVResp
	}
}

func (pd *ParameterDecisions) AddObjectResponseParser() acsxml.AddObjectResponseStruct {
	var addObject acsxml.AddObjectResponseStruct
	_ = xml.Unmarshal(pd.ReqRes.Body, &addObject)

	return addObject
}
