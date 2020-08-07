package ngapConvert

import (
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/lib/openapi/models"
)

func AllowedNssaiToNgap(allowedNssaiModels []models.AllowedSnssai) (allowedNssaiNgap ngapType.AllowedNSSAI) {
	for _, allowedSnssai := range allowedNssaiModels {
		item := ngapType.AllowedNSSAIItem{
			SNSSAI: SNssaiToNgap(*allowedSnssai.AllowedSnssai),
		}
		allowedNssaiNgap.List = append(allowedNssaiNgap.List, item)
	}
	return
}

func AllowedNssaiToModels(allowedNssaiNgap ngapType.AllowedNSSAI) (allowedNssaiModels []models.AllowedSnssai) {
	for _, item := range allowedNssaiNgap.List {
		snssai := SNssaiToModels(item.SNSSAI)
		allowedSnssai := models.AllowedSnssai{
			AllowedSnssai: &snssai,
		}
		allowedNssaiModels = append(allowedNssaiModels, allowedSnssai)
	}
	return
}
