package ngapConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/lib/openapi/models"
	"strings"
)

func PlmnIdToModels(ngapPlmnId ngapType.PLMNIdentity) (modelsPlmnid models.PlmnId) {
	value := ngapPlmnId.Value
	hexString := strings.Split(hex.EncodeToString(value), "")
	modelsPlmnid.Mcc = hexString[1] + hexString[0] + hexString[3]
	if hexString[2] == "f" {
		modelsPlmnid.Mnc = hexString[5] + hexString[4]
	} else {
		modelsPlmnid.Mnc = hexString[2] + hexString[5] + hexString[4]
	}
	return
}
func PlmnIdToNgap(modelsPlmnid models.PlmnId) (ngapPlmnId ngapType.PLMNIdentity) {
	var hexString string
	mcc := strings.Split(modelsPlmnid.Mcc, "")
	mnc := strings.Split(modelsPlmnid.Mnc, "")
	if len(modelsPlmnid.Mnc) == 2 {
		hexString = mcc[1] + mcc[0] + "f" + mcc[2] + mnc[1] + mnc[0]
	} else {
		hexString = mcc[1] + mcc[0] + mnc[0] + mcc[2] + mnc[2] + mnc[1]
	}
	ngapPlmnId.Value, _ = hex.DecodeString(hexString)
	return
}
