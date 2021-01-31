package types

import (
	"log"
	"testing"
)

func TestIsObjectParameter(t *testing.T) {
	params := []ParameterValueStruct{
		{Name: "InternetGatewayDevice.LANDevice.", Flag: Flag{
			Read:  true,
			Write: false,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.", Flag: Flag{
			Read:  true,
			Write: false,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.", Flag: Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.", Flag: Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Enable", Flag: Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.", Flag: Flag{
			Read:  true,
			Write: false,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey", Flag: Flag{
			Read:  true,
			Write: true,
		}},
	}

	results := map[int]bool{
		0: false,
		1: false,
		2: true,
		3: false,
		4: false,
		5: false,
		6: false,
	}

	for idx, param := range params {
		if IsObjectParameter(param.Name, param.Flag.Write) != results[idx] {
			t.Error("Param", param.Name, "fails IsObjectParameter")
		}
	}
}

func TestObjectParamToInstance(t *testing.T) {
	log.Println(ObjectParamToInstance("InternetGatewayDevice.LANDevice.1.WLANConfiguration.1."))
}
