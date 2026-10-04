/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package types

import (
	"encoding/hex"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	"my5G-RANTester/test/aio5gc/lib/convert"
)

type Tai struct {
	Tac            string
	plmnSnssaiList []models.Nrf_NFMgmt_PlmnSnssai
}

func TaiListToModels(list ie.SupportedTAList) []Tai {
	var out []Tai
	for _, item := range list.List {
		tai := Tai{Tac: hex.EncodeToString(item.TAC.Value)}
		for _, broadcast := range item.BroadcastPLMNList.List {
			plmn := convert.PLMNToModels(broadcast.PLMNIdentity)
			record := models.Nrf_NFMgmt_PlmnSnssai{PlmnId: &plmn}
			for _, slice := range broadcast.TAISliceSupportList.List {
				s := models.ExtSnssai{Sst: int32(slice.SNSSAI.SST.Value[0])}
				if slice.SNSSAI.SD != nil {
					s.Sd = hex.EncodeToString(slice.SNSSAI.SD.Value)
				}
				record.SNssaiList = append(record.SNssaiList, s)
			}
			tai.plmnSnssaiList = append(tai.plmnSnssaiList, record)
		}
		out = append(out, tai)
	}
	return out
}
