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
	"my5G-RANTester/test/aio5gc/msg/nas"

	"github.com/free5gc/openapi/models"
)

func InitialUEMessage(req *ngapType.InitialUEMessage, gnb *context.GNBContext, fgc *context.Aio5gc) error {

	amf := fgc.GetAMFContext()
	ue := &context.UEContext{}
	var nasMsg *ngapType.NASPDU

	nrLocation := models.NrLocation{}
	var ueRanNgapId int64
	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ueRanNgapId = req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value
			ue = amf.NewUE(ueRanNgapId)

		case ngapType.ProtocolIEIDNASPDU:
			nasMsg = req.ProtocolIEs.List[ie].Value.NASPDU

		case ngapType.ProtocolIEIDUserLocationInformation:
			UserLocationInformationNR := req.ProtocolIEs.List[ie].Value.UserLocationInformation.UserLocationInformationNR
			tai := ngapConvert.TaiToModels(UserLocationInformationNR.TAI)
			nrLocation.Tai = &tai
			ncgi := models.Ncgi{}
			ncgi.NrCellId = ngapConvert.BitStringToHex(&UserLocationInformationNR.NRCGI.NRCellIdentity.Value)
			plmn := ngapConvert.PlmnIdToModels(UserLocationInformationNR.NRCGI.PLMNIdentity)
			ncgi.PlmnId = &plmn
			nrLocation.Ncgi = &ncgi
			nrLocation.GlobalGnbId = gnb.GetGlobalRanNodeID()

		case ngapType.ProtocolIEIDRRCEstablishmentCause:

		case ngapType.ProtocolIEIDUEContextRequest:

		default:
			return errors.New("[5GC][NGAP] Received unknown ie for InitialUEMessage")
		}
	}
	ue.SetUserLocationInfo(&nrLocation)

	nas.Dispatch(nasMsg, ue, fgc, gnb)
	return nil
}
