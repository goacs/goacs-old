package methods

import (
	"goacs/acs/types"
	"sort"
	"testing"
)

func TestParameterDecisions_Sorting_Parameters(t *testing.T) {
	params := []types.ParameterInfo{
		{Name: "InternetGatewayDevice.LANDevice.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Enable", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey", Writable: "1", Done: false},
	}

	sort.Sort(sort.Reverse(types.SortParamsInfo(params)))

	if params[len(params)-1].Name != "InternetGatewayDevice.LANDevice." {
		t.Error("Last param are not InternetGatewayDevice.LANDevice.")
	}

	if params[0].Name != "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey" {
		t.Error("First param are not InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey")
	}
}

func TestParameterDecisions_GetNextLevelParams(t *testing.T) {
	params := []types.ParameterInfo{
		{Name: "InternetGatewayDevice.LANDevice.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Enable", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.Status", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.1.WEPKey", Writable: "1", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.2.", Writable: "0", Done: false},
		{Name: "InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.WEPKey.2.WEPKey", Writable: "1", Done: false},
	}

	sort.Sort(sort.Reverse(types.SortParamsInfo(params)))
}
