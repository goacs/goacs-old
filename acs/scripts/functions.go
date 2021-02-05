package scripts

import (
	"goacs/acs/types"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
	"strings"
)

func (se *ScriptEngine) SetParameter(path string, value string, flags string) {
	flag, _ := types.FlagFromString(flags)
	parameter := types.ParameterValueStruct{
		Name: path,
		ValueStruct: types.ValueStruct{
			Value: value,
			Type:  "",
		},
		Flag: flag,
	}

	currentParameter := se.ReqRes.Session.CPE.GetParameter(path)

	if currentParameter != nil {
		parameter.ValueStruct.Type = currentParameter.ValueStruct.Type
	}

	se.ReqRes.Session.CPE.AddParameter(parameter)
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_, _ = cpeRepository.UpdateParameter(&se.ReqRes.Session.CPE, parameter)
	if flag.System == false {
		//TODO Tricky place, we have method to add Param..
		se.ReqRes.Session.ParametersToAdd = append(se.ReqRes.Session.ParametersToAdd, parameter)
	}
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

func (se *ScriptEngine) ParameterExist(path string) bool {
	return se.ReqRes.Session.CPE.ParameterValueExist(path)
}

func (se *ScriptEngine) SaveDevice() {
	cpeRepository := mysql.NewCPERepository(repository.GetConnection())
	_ = cpeRepository.BulkInsertOrUpdateParameters(&se.ReqRes.Session.CPE, se.ReqRes.Session.CPE.ParameterValues)
}

func (se *ScriptEngine) Download(filename string, filetype string) {
	dlTask := tasks.NewCPETask(se.ReqRes.Session.CPE.UUID)
	dlTask.AsUploadFirmware(filename, filetype)
	se.ReqRes.Session.AddTask(dlTask)
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
