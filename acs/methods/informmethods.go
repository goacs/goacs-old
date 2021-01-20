package methods

import (
	"encoding/xml"
	"goacs/acs/http"
	acsxml "goacs/acs/types"
	"goacs/lib"
	"goacs/models/tasks"
	"goacs/repository/mysql"
	"log"
)

type InformDecision struct {
	ReqRes *http.CPERequest
}

func (InformDecision *InformDecision) CpeInformResponse() string {
	InformDecision.ReqRes.Session.PrevReqType = acsxml.InformReq
	return InformDecision.ReqRes.Envelope.InformResponse()
}

func (InformDecision *InformDecision) CpeInformRequestParser() {
	env := new(lib.Env)
	var inform acsxml.Inform
	_ = xml.Unmarshal(InformDecision.ReqRes.Body, &inform)
	log.Println("SESSION FROM InformReq", InformDecision.ReqRes.Session.IsNew, InformDecision.ReqRes.Session.ReadAllParameters)

	InformDecision.ReqRes.Session.FillCPESessionFromInform(inform)
	cpeRepository := mysql.NewCPERepository(InformDecision.ReqRes.DBConnection)
	_, cpeExist, _ := cpeRepository.UpdateOrCreate(&InformDecision.ReqRes.Session.CPE)
	InformDecision.ReqRes.Session.ReadAllParameters = !cpeExist
	InformDecision.ReqRes.Session.IsNewInACS = !cpeExist
	InformDecision.ReqRes.Session.Provision = !cpeExist
	InformDecision.ReqRes.Session.IsNew = false

	if env.Get("DEBUG", "false") == "true" {
		InformDecision.ReqRes.Session.IsBoot = true
	}

	_, _ = cpeRepository.SaveParameters(&InformDecision.ReqRes.Session.CPE)
	task := tasks.NewCPETask(InformDecision.ReqRes.Session.CPE.UUID)
	task.Task = acsxml.InformResp
	InformDecision.ReqRes.Session.AddTask(task)

	if InformDecision.ReqRes.Session.IsNewInACS || InformDecision.ReqRes.Session.IsBoot {
		task = tasks.NewCPETask(InformDecision.ReqRes.Session.CPE.UUID)
		task.Task = acsxml.GPNReq
		task.ParameterInfo = append(task.ParameterInfo, acsxml.ParameterInfo{
			Name: InformDecision.ReqRes.Session.CPE.Root + ".",
			Done: false,
		})
		InformDecision.ReqRes.Session.AddTask(task)
	}

}
