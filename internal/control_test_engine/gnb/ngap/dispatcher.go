/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package ngap

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	ngapType "github.com/free5gc/ngap/ie"
	ngapmsg "github.com/free5gc/ngap/message"

	log "github.com/sirupsen/logrus"
)

// Dispatch decodes a message read from an AMF's association and handles it. The
// messages of a UE are handled in the order they were received, on the UE's own queue:
// handled each in its own goroutine, a UE Context Release Command could overtake the
// NAS message sent before it. A UE slow to take its messages delays no other UE. The
// other messages are handled here, before the next one is read.
func Dispatch(amf *context.GNBAmf, gnb *context.GNBContext, payload []byte) {
	msg, err := ngapmsg.Parse(payload)
	if err != nil || msg == nil {
		log.Errorf("[GNB][NGAP] Unable to decode message for gNB %s: %v", gnb.GetGnbId(), err)
		return
	}
	if ue := messageUE(gnb, msg); ue != nil {
		ue.Enqueue(func() { handleMessage(amf, gnb, msg) })
	} else if _, update := msg.(*ngapmsg.AMFConfigurationUpdate); update {
		// It dials the AMFs it adds, which an unreachable one would hold up for seconds.
		go handleMessage(amf, gnb, msg)
	} else {
		handleMessage(amf, gnb, msg)
	}
}

// messageUE returns the UE a message is for: the one with its RAN UE NGAP ID, or else
// the one with its AMF UE NGAP ID. The AMF's ID is recorded here, as soon as it is
// received, so that a following message with that ID alone finds the UE even if this
// one is not handled yet. A Handover Request is for no UE yet, it creates one: handled
// by Dispatch itself, its UE exists before the next message is read.
func messageUE(gnb *context.GNBContext, message ngapmsg.Message) *context.GNBUe {
	var ranID *ngapType.RANUENGAPID
	var amfID *ngapType.AMFUENGAPID
	switch value := message.(type) {
	case *ngapmsg.DownlinkNASTransport:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.InitialContextSetupRequest:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PDUSessionResourceSetupRequest:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PDUSessionResourceReleaseCommand:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.ErrorIndication:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PathSwitchRequestAcknowledge:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.HandoverCommand:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.UEContextReleaseCommand:
		if value.UENGAPIDs != nil {
			switch ids := value.UENGAPIDs.Choice.(type) {
			case *ngapType.UENGAPIDPair:
				ranID, amfID = ids.RANUENGAPID, ids.AMFUENGAPID
			case *ngapType.AMFUENGAPID:
				amfID = ids
			}
		}
	}
	var ue *context.GNBUe
	if ranID != nil {
		ue, _ = gnb.GetGnbUe(ranID.Value)
	}
	if amfID != nil {
		if ue == nil {
			ue, _ = gnb.GetGnbUeByAmfUeId(amfID.Value)
		} else {
			ue.SetAmfUeId(amfID.Value)
		}
	}
	return ue
}

func handleMessage(amf *context.GNBAmf, gnb *context.GNBContext, msg ngapmsg.Message) {
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
