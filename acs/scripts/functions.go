package scripts

import (
	"fmt"
	"goacs/acs/methods"
	"goacs/acs/types"
	"goacs/models/cpe"
	"goacs/repository"
	"goacs/repository/mysql"
	"strings"
)

func (se *ScriptEngine) SetParameter(path string, value string, flags string) {
	flag, _ := types.FlagFromString(flags)
	se.ReqRes.Session.CPE.AddParameter(types.ParameterValueStruct{
		Name:  path,
		Value: value,
		Type:  "",
		Flag:  flag,
	})
}

func (se *ScriptEngine) GetParameterValue(path string) string {
	if value, err := se.ReqRes.Session.CPE.GetParameterValue(path); err == nil {
		return value
	}

	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	cpeParameters, _ := cpeRepository.GetCPEParameters(&se.ReqRes.Session.CPE)
	se.ReqRes.Session.CPE.AddParameterValues(cpeParameters)

	value, err := se.ReqRes.Session.CPE.GetParameterValue(path)

	if err != nil {
		return ""
	}

	return value
}

func (se *ScriptEngine) SaveDevice() {
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_ = cpeRepository.BulkInsertOrUpdateParameters(&se.ReqRes.Session.CPE, se.ReqRes.Session.CPE.ParameterValues)
}

func (se *ScriptEngine) AddObject(path string) {
	reqBody := se.ReqRes.Envelope.AddObjectRequest(path, "")
	_, _ = fmt.Fprint(se.ReqRes.Response, reqBody)
}

func (se *ScriptEngine) SendParameters() {
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	templateRepository := mysql.NewTemplateRepository(repository.GetConnection())
	allParameters, _ := cpeRepository.GetCPEParameters(&se.ReqRes.Session.CPE)
	templateParameters := templateRepository.GetPrioritizedParametersForCPE(&se.ReqRes.Session.CPE)
	writableParameters := types.GetParametersWithFlag(allParameters, "S")
	writableParameters = cpe.CombineTemplateParameters(writableParameters, templateParameters)
	se.ReqRes.Session.CPE.ParametersQueue = writableParameters
	parameterDecisions := methods.ParameterDecisions{ReqRes: se.ReqRes}
	parameterDecisions.SetParameterValuesRequest()
}

func (se *ScriptEngine) StringContains(text string, search string) bool {
	return strings.Contains(text, search)
}

func (se *ScriptEngine) SubString(text string, start int, end int) string {
	return text[start:end]
}

func (se *ScriptEngine) Replace(text string, from string, to string) string {
	return strings.ReplaceAll(text, from, to)
}
