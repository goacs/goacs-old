package cpe

import (
	"goacs/acs/types"
	"log"
	"testing"
)

//func GetParameterWithFlagTest(t *testing.T) {
//	cpe := CPE{}
//	cpe.AddParameter(types.ParameterValueStruct{
//		Name:
//	})
//}

func TestCompareObjectParameters(t *testing.T) {
	cpeParams := []types.ParameterValueStruct{
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Enable", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.Enable", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.WEPKey.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.WEPKey.1.WEPKey", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
	}

	dbParams := []types.ParameterValueStruct{
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Enable", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.Enable", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.WEPKey.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.2.WEPKey.1.WEPKey", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.3.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.3.Enable", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.3.WEPKey.1.", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.3.WEPKey.1.WEPKey", Flag: types.Flag{
			Read:  true,
			Write: true,
		}},
	}

	addparams, delparams := CompareObjectParameters(cpeParams, dbParams)

	log.Println(addparams)
	log.Println(delparams)
	//assert.Len(t, addparams, 1)
	//assert.Len(t, delparams, 0)
}
