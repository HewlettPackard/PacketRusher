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

func UplinkNASTransport(req *ngapType.UplinkNASTransport, gnb *context.GNBContext, fgc *context.Aio5gc) error {

	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error
	var naspdu *ngapType.NASPDU
	nrLocation := models.NrLocation{}
	amf := fgc.GetAMFContext()

	for ie := range req.ProtocolIEs.List {
		switch req.ProtocolIEs.List[ie].Id.Value {
		case ngapType.ProtocolIEIDRANUENGAPID:
			ranUe, err = amf.FindUEByRanId(req.ProtocolIEs.List[ie].Value.RANUENGAPID.Value)
			if err != nil {
				return err
			}
		case ngapType.ProtocolIEIDAMFUENGAPID:
			ue, err = amf.FindUEById(req.ProtocolIEs.List[ie].Value.AMFUENGAPID.Value)
			if err != nil {
				return err
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
			return errors.New("[5GC][NGAP] Received unknown ie for UplinkNASTransport")
		}
	}
	if ue != ranUe {
		return errors.New("[5GC][NGAP] RanUeNgapId does not match the one registred for this UE")
	}

	nas.Dispatch(naspdu, ue, fgc, gnb)
	return nil
}
