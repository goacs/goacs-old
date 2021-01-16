package cpe

import (
	"errors"
	"goacs/acs/types"
	"log"
	"strings"
	"time"
)

type CPE struct {
	UUID                      string          `json:"uuid" db:"uuid"`
	SerialNumber              string          `json:"serial_number" db:"serial_number"`
	OUI                       string          `json:"oui" db:"oui"`
	ProductClass              string          `json:"product_class" db:"product_class"`
	Manufacturer              string          `json:"manufacturer" db:"manufacturer"`
	SoftwareVersion           string          `json:"software_version" db:"software_version"`
	HardwareVersion           string          `json:"hardware_version" db:"hardware_version"`
	IpAddress                 types.IPAddress `json:"ip_address" db:"ip_address"`
	ConnectionRequestUser     string          `json:"connection_request_user" db:"connection_request_user"`
	ConnectionRequestPassword string          `json:"connection_request_password" db:"connection_request_password"`
	ConnectionRequestUrl      string          `json:"connection_request_url" db:"connection_request_url"`
	Root                      string
	ParametersInfo            []types.ParameterInfo
	ParameterValues           []types.ParameterValueStruct
	ParametersQueue           []types.ParameterValueStruct
	Fault                     types.Fault
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

func (cpe *CPE) AddParameterInfo(parameter types.ParameterInfo) {
	for index := range cpe.ParametersInfo {
		if cpe.ParametersInfo[index].Name == parameter.Name {
			cpe.ParametersInfo[index].Done = parameter.Done
			cpe.ParametersInfo[index].Writable = parameter.Writable
			return
		}
	}

	cpe.ParametersInfo = append(cpe.ParametersInfo, parameter)
}

func (cpe *CPE) AddParametersInfo(parameters []types.ParameterInfo) {
	for _, parameter := range parameters {
		cpe.AddParameterInfo(parameter)
	}
}

func (cpe *CPE) GetParameterInfoByName(name string) (types.ParameterInfo, error) {
	for _, parameter := range cpe.ParametersInfo {
		if parameter.Name == name {
			return parameter, nil
		}
	}

	return types.ParameterInfo{}, errors.New("Cannot find parameter")
}

func (cpe *CPE) UpdateParameterFlags(parameterName string, flag types.Flag) {
	for index := range cpe.ParameterValues {
		if cpe.ParameterValues[index].Name == parameterName {
			//Replace exist parameter
			cpe.ParameterValues[index].Flag = flag
			return
		}
	}
}

func (cpe *CPE) AddParameter(parameter types.ParameterValueStruct) {
	parameterInfo, err := cpe.GetParameterInfoByName(parameter.Name)

	if err == nil {
		if parameterInfo.Writable == "1" {
			parameter.Flag.Write = true
		}
	}

	for index := range cpe.ParameterValues {
		if cpe.ParameterValues[index].Name == parameter.Name {
			//Replace exist parameter
			cpe.ParameterValues[index].ValueStruct.Value = parameter.ValueStruct.Value
			if parameter.ValueStruct.Type != "" {
				cpe.ParameterValues[index].ValueStruct.Type = parameter.ValueStruct.Type
			}
			cpe.ParameterValues[index].Flag = parameter.Flag
			return
		}
	}

	cpe.ParameterValues = append(cpe.ParameterValues, parameter)
}

func (cpe *CPE) AddParameterValues(parameters []types.ParameterValueStruct) {
	for _, parameter := range parameters {
		parameter.Flag.Read = true
		cpe.AddParameter(parameter)
	}
}

func CombineTemplateParameters(cpeParameters []types.ParameterValueStruct, templateParameters []types.PrioritizedParameters) []types.ParameterValueStruct {
	var parameters []types.ParameterValueStruct
	for _, cpeParameter := range cpeParameters {
		parameterExist := false
		for _, templateParameter := range templateParameters {
			if cpeParameter.Name == templateParameter.Name {
				if templateParameter.Priority > 100 {
					parameterExist = true
					parameters = append(parameters, templateParameter.ParameterValueStruct)
				}
			}
		}

		if parameterExist == false {
			parameters = append(parameters, cpeParameter)
		}
	}

	for _, templateParameter := range templateParameters {
		for _, cpeParameter := range cpeParameters {
			if cpeParameter.Name == templateParameter.Name {
				if templateParameter.Priority > 100 {
					continue
				}
			}
		}
		parameters = append(parameters, templateParameter.ParameterValueStruct)
	}

	return parameters
}

func (cpe *CPE) ApplyTemplateParameters(parameters []types.PrioritizedParameters) {
	result := CombineTemplateParameters(cpe.ParameterValues, parameters)
	cpe.AddParameterValues(result)
}

func (cpe *CPE) ParameterValueExist(parameterName string) bool {
	for _, parameterValue := range cpe.ParameterValues {
		if parameterValue.Name == parameterName {
			return true
		}
	}

	return false
}

func (cpe *CPE) GetParameter(parameterName string) *types.ParameterValueStruct {
	for _, parameter := range cpe.ParameterValues {
		if parameter.Name == parameterName {
			return &parameter
		}
	}

	return nil
}

func (cpe *CPE) GetParameterValue(parameterName string) (string, error) {
	for index := range cpe.ParameterValues {
		if cpe.ParameterValues[index].Name == parameterName {
			return cpe.ParameterValues[index].ValueStruct.Value, nil
		}
	}

	return "", errors.New("Unable to find parameter " + parameterName + " in CPE")
}

func (cpe *CPE) GetAddObjectParameters() []types.ParameterValueStruct {
	var filteredParameters []types.ParameterValueStruct
	for _, parameter := range cpe.ParametersInfo {
		// If Last character of parameter name is ".", then add it as AddObject to DB
		if parameter.Name[len(parameter.Name)-1:] == "." && parameter.Writable == "1" {
			filteredParameters = append(filteredParameters, types.ParameterValueStruct{
				Name: parameter.Name,
				ValueStruct: types.ValueStruct{
					Value: "",
					Type:  "",
				},
				Flag: types.Flag{
					Read:         true,
					Write:        true,
					AddObject:    true,
					System:       false,
					PeriodicRead: false,
					Important:    false,
				},
			})
		}
	}

	return filteredParameters
}

func (cpe *CPE) GetFullPathParameterNames() []types.ParameterInfo {
	var filteredParameters []types.ParameterInfo
	for _, parameter := range cpe.ParametersInfo {
		//check if last char in Name is not equal to . (dot)
		if parameter.Name[len(parameter.Name)-1:] != "." {
			filteredParameters = append(filteredParameters, parameter)
		}
	}

	return filteredParameters
}
func (cpe *CPE) PopParametersQueue() []types.ParameterValueStruct {

	defer func() {
		cpe.ParametersQueue = []types.ParameterValueStruct{}
	}()

	return cpe.ParametersQueue
}

func (cpe *CPE) GetParametersWithFlag(flag string) []types.ParameterValueStruct {
	return types.GetParametersWithFlag(cpe.ParameterValues, flag)
}

func (cpe *CPE) SetRoot(root string) {
	if root == "Device" || root == "InternetGatewayDevice" {
		cpe.Root = root
	}
}

func (cpe *CPE) Fails() bool {
	return cpe.Fault.FaultCode != "" || cpe.Fault.FaultString != ""
}

func (cpe *CPE) GetChangedParametersToWrite(otherParameters *[]types.ParameterValueStruct) []types.ParameterValueStruct {
	parametersDiff := []types.ParameterValueStruct{}

	for _, cpeParam := range cpe.ParameterValues {
		for _, otherParam := range *otherParameters {
			if otherParam.Flag.Write == true && otherParam.Name == cpeParam.Name && otherParam.ValueStruct.Value != cpeParam.ValueStruct.Value {
				log.Println("other param", otherParam.ValueStruct.Value)
				log.Println("cpe param", cpeParam.ValueStruct.Value)
				parametersDiff = append(parametersDiff, otherParam)
				break
			}
		}
	}

	return parametersDiff
}

func DetermineDeviceTreeRootPath(parameters []types.ParameterValueStruct) string {
	for _, parameter := range parameters {
		splittedParamName := strings.Split(parameter.Name, ".")

		if splittedParamName[0] == "Device" {
			return "Device"
		}
	}

	return "InternetGatewayDevice"
}
