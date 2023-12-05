/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/types"
	"my5G-RANTester/test/aio5gc/msg"
)

func NGSetupRequest(req *ngapType.NGSetupRequest, gnb *context.GNBContext, fgc *context.Aio5gc) (err error) {
	// assert req contains wanted values?
	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDGlobalRANNodeID:
			globalRANNodeID := ngapConvert.RanIdToModels(*req.ProtocolIEs.List[ie].Value.GlobalRANNodeID)
			gnb.SetGlobalRanNodeID(globalRANNodeID)
		case ngapType.ProtocolIEIDRANNodeName:
			gnb.SetRanNodename(req.ProtocolIEs.List[ie].Value.RANNodeName.Value)
		case ngapType.ProtocolIEIDSupportedTAList:
			supportedTaiList := types.TaiListToModels(*req.ProtocolIEs.List[ie].Value.SupportedTAList)
			gnb.SetSuportedTAList(supportedTaiList)
		case ngapType.ProtocolIEIDDefaultPagingDRX:
			gnb.SetDefautlPagingDRX(*req.ProtocolIEs.List[ie].Value.DefaultPagingDRX)
		default:
			return errors.New("[5GC][NGAP] Received unknown ie for NGSetupRequest")
		}
	}

	msg.SendNGSetupResponse(gnb, fgc)

	return nil
}
