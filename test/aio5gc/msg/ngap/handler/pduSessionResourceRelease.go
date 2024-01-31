/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"my5G-RANTester/test/aio5gc/context"

	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func PDUSessionResourceRelease(req *ngapType.PDUSessionResourceReleaseResponse, fgc *context.Aio5gc) error {

	amf := fgc.GetAMFContext()
	var ue *context.UEContext
	var ranUe *context.UEContext
	var err error
	var releasedPdus []ngapType.PDUSessionResourceReleasedItemRelRes

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
		case ngapType.ProtocolIEIDPDUSessionResourceReleasedListRelRes:
			releasedPdus = req.ProtocolIEs.List[ie].Value.PDUSessionResourceReleasedListRelRes.List
		}
	}
	if ue != ranUe {
		return errors.New("[5GC][NGAP] RanUeNgapId does not match the one Registered for this UE")
	}
	for i := range releasedPdus {
		pduSessionID := int32(releasedPdus[i].PDUSessionID.Value)
		err := context.ConfirmPDUSessionRelease(ue, pduSessionID)
		if err != nil {
			log.Errorf("[5GC][NGAP] Error in PDU session resource release response handle: " + err.Error())
		}
	}
	return nil
}
