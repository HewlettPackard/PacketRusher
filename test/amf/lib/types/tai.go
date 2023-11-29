/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package types

import (
	"encoding/hex"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"

	"github.com/free5gc/openapi/models"
)

type Tai struct {
	Tac            string
	plmnSnssaiList []models.PlmnSnssai
}

func TaiListToModels(taiList ngapType.SupportedTAList) []Tai {
	taiModels := []Tai{}
	for i := range taiList.List {
		taiModel := Tai{}
		taiModel.Tac = hex.EncodeToString(taiList.List[i].TAC.Value)
		for j := range taiList.List[i].BroadcastPLMNList.List {
			plmnSnssai := models.PlmnSnssai{}
			plmnid := ngapConvert.PlmnIdToModels(taiList.List[i].BroadcastPLMNList.List[j].PLMNIdentity)
			plmnSnssai.PlmnId = &plmnid

			for k := range taiList.List[i].BroadcastPLMNList.List[j].TAISliceSupportList.List {
				SNssai := taiList.List[i].BroadcastPLMNList.List[j].TAISliceSupportList.List[k].SNSSAI
				plmnSnssai.SNssaiList = append(plmnSnssai.SNssaiList, ngapConvert.SNssaiToModels(SNssai))
			}
			taiModel.plmnSnssaiList = append(taiModel.plmnSnssaiList, plmnSnssai)
		}
		taiModels = append(taiModels, taiModel)
	}
	return taiModels
}
