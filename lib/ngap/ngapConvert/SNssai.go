package ngapConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/lib/openapi/models"
)

func SNssaiToModels(ngapSnssai ngapType.SNSSAI) (modelsSnssai models.Snssai) {
	modelsSnssai.Sst = int32(ngapSnssai.SST.Value[0])
	if ngapSnssai.SD != nil {
		modelsSnssai.Sd = hex.EncodeToString(ngapSnssai.SD.Value)
	}
	return
}

func SNssaiToNgap(modelsSnssai models.Snssai) (ngapSnssai ngapType.SNSSAI) {
	ngapSnssai.SST.Value = []byte{byte(modelsSnssai.Sst)}

	if modelsSnssai.Sd != "" {
		ngapSnssai.SD = new(ngapType.SD)
		ngapSnssai.SD.Value, _ = hex.DecodeString(modelsSnssai.Sd)
	}
	return
}
