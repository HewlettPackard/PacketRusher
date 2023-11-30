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

func UplinkNASTransport(req *ngapType.UplinkNASTransport, amf context.AMFContext, gnb context.GNBContext, nasHandler func(*ngapType.NASPDU, *context.UEContext) ([]byte, int64, error)) ([]byte, error) {

	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error
	var naspdu *ngapType.NASPDU
	var msg []byte
	nrLocation := models.NrLocation{}

	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ranUe, err = amf.FindUEByRanId(req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value)
			if err != nil {
				return msg, err
			}
		case ngapType.ProtocolIEIDAMFUENGAPID:
			ue, err = amf.FindUEById(req.ProtocolIEs.List[ie].Value.AMFUENGAPID.Value)
			if err != nil {
				return msg, err
			}

		case ngapType.ProtocolIEIDNASPDU:
			naspdu = req.ProtocolIEs.List[ie].Value.NASPDU

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
			return msg, errors.New("[AMF][NGAP] Received unknown ie for UplinkNASTransport")
		}
	}
	if ue != ranUe {
		return msg, errors.New("[AMF][NGAP] RanUeNgapId does not match the one registred for this UE")
	}
	resNasMsg, resNgapType, err := nasHandler(naspdu, ue)
	if err != nil {
		return msg, err
	}

	var resType string
	switch resNgapType {
	case ngapType.ProcedureCodeInitialContextSetup:
		msg, err = InitialContextSetupRequest(resNasMsg, ue, amf)
		resType = "Initial Context Setup Request"

	case ngapType.ProcedureCodeDownlinkNASTransport:
		msg, err = DownlinkNASTransport(resNasMsg, ue)
		resType = "Downlink NAS Transport"

	case ngapType.ProcedureCodePDUSessionResourceSetup:
		// msg, err = PDUSessionResourceSetup(resNasMsg, ue, amf)
		// resType = "Downlink NAS Transport"
		err = errors.New("PDU Session Resource Setup not implemented yet")

	case 0:
		return nil, nil
	}
	if err != nil {
		return msg, err
	}
	log.Info("[AMF][NGAP] Sent " + resType)

	return msg, nil
}
