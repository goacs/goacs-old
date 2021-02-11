package methods

import (
	"goacs/acs/http"
	acsxml "goacs/acs/types"
	"goacs/repository/mysql"
	"log"
)

type FaultDecision struct {
	ReqRes *http.CPERequest
}

func (FaultDecision *FaultDecision) ResponseDecision() {
	log.Print(string(FaultDecision.ReqRes.Body))
	FaultDecision.ReqRes.Session.PrevState = acsxml.FaultResp
	repository := mysql.NewFaultRepository()
	repository.SaveFault(&FaultDecision.ReqRes.Session.CPE,
		FaultDecision.ReqRes.Session.CPE.Fault.DetailFaultCode,
		FaultDecision.ReqRes.Session.CPE.Fault.DetailFaultString,
	)
}
