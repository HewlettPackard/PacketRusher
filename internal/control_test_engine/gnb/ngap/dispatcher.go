/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	ngapmsg "github.com/free5gc/ngap/message"

	log "github.com/sirupsen/logrus"
)

func Dispatch(amf *context.GNBAmf, gnb *context.GNBContext, payload []byte) {
	msg, err := ngapmsg.Parse(payload)
	if err != nil || msg == nil {
		log.Errorf("[GNB][NGAP] Unable to decode message for gNB %s: %v", gnb.GetGnbId(), err)
		return
	}
	dispatchUEMessage(amf, gnb, msg)
}

func dispatchUEMessage(amf *context.GNBAmf, gnb *context.GNBContext, message ngapmsg.Message) {
	ran, amfID, hasAMF := messageUEIDs(message)
	var ue *context.GNBUe
	if ran != 0 {
		ue, _ = gnb.GetGnbUe(ran)
	}
	if ue == nil && hasAMF {
		ue, _ = gnb.GetGnbUeByAmfUeId(amfID)
	}
	if ue != nil {
		ue.ProcessDownlink(func() { dispatchDecoded(amf, gnb, message) })
	} else {
		dispatchDecoded(amf, gnb, message)
	}
}

func dispatchDecoded(amf *context.GNBAmf, gnb *context.GNBContext, msg ngapmsg.Message) {
	switch value := msg.(type) {
	case *ngapmsg.DownlinkNASTransport:
		HandlerDownlinkNasTransport(gnb, value)
	case *ngapmsg.InitialContextSetupRequest:
		HandlerInitialContextSetupRequest(gnb, value)
	case *ngapmsg.PDUSessionResourceSetupRequest:
		HandlerPduSessionResourceSetupRequest(gnb, value)
	case *ngapmsg.PDUSessionResourceReleaseCommand:
		HandlerPduSessionReleaseCommand(gnb, value)
	case *ngapmsg.UEContextReleaseCommand:
		HandlerUeContextReleaseCommand(gnb, value)
	case *ngapmsg.AMFConfigurationUpdate:
		HandlerAmfConfigurationUpdate(amf, gnb, value)
	case *ngapmsg.AMFStatusIndication:
		HandlerAmfStatusIndication(amf, gnb, value)
	case *ngapmsg.HandoverRequest:
		HandlerHandoverRequest(amf, gnb, value)
	case *ngapmsg.Paging:
		HandlerPaging(gnb, value)
	case *ngapmsg.ErrorIndication:
		HandlerErrorIndication(gnb, value)
	case *ngapmsg.NGSetupResponse:
		HandlerNgSetupResponse(amf, gnb, value)
	case *ngapmsg.PathSwitchRequestAcknowledge:
		HandlerPathSwitchRequestAcknowledge(gnb, value)
	case *ngapmsg.HandoverCommand:
		HandlerHandoverCommand(amf, gnb, value)
	case *ngapmsg.NGSetupFailure:
		HandlerNgSetupFailure(amf, gnb, value)
	default:
		log.Warnf("[GNB][NGAP] Unhandled message %T", msg)
	}
}
