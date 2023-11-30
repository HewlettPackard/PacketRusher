/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngapMsgHandler

import (
	"errors"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"

	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
)

func InitialUEMessage(req *ngapType.InitialUEMessage, amf *context.AMFContext, gnb context.GNBContext, nasHandler func(*ngapType.NASPDU, *context.UEContext) ([]byte, int64, error)) (msg []byte, err error) {
	// assert req contains wanted values?

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
			return msg, errors.New("[AMF][NGAP] Received unknown ie for InitialUEMessage")
		}
	}
	ue.SetUserLocationInfo(&nrLocation)

	nasRes, _, err := nasHandler(nasMsg, ue)
	if err != nil {
		return msg, err
	}

	msg, err = DownlinkNASTransport(nasRes, ue)
	if err != nil {
		return msg, err
	}
	log.Info("[AMF][NGAP] Sent Authentication Request")
	return msg, nil
}
