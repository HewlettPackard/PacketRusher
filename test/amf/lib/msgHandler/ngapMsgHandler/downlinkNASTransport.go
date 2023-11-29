/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngapMsgHandler

import (
	"errors"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"
	"my5G-RANTester/test/amf/lib/msgHandler/nasMsgHandler"

	"github.com/free5gc/nas"
)

func DownlinkNASTransport(messageType uint8, ue *context.UEContext) ([]byte, error) {

	nasPdu := []byte{}
	var err error

	switch messageType {
	case nas.MsgTypeAuthenticationRequest:
		nasPdu, err = nasMsgHandler.AuthenticationRequest(ue)
	case nas.MsgTypeSecurityModeCommand:
		nasPdu, err = nasMsgHandler.SecurityModeCommand(ue)
	case nas.MsgTypePDUSessionEstablishmentAccept:
		nasPdu, err = nasMsgHandler.PDUSessionEstablishmentAccept(ue)
	}
	if err != nil {
		return nasPdu, errors.New("[AMF] Error creating nas message: " + err.Error())
	}

	message, err := BuildDownlinkNASTransport(nasPdu, *ue)
	if err != nil {
		return []byte{}, err
	}
	return ngap.Encoder(message)
}

func BuildDownlinkNASTransport(nasPdu []byte, ue context.UEContext) (pdu ngapType.NGAPPDU, err error) {

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeDownlinkNASTransport
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentIgnore

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentDownlinkNASTransport
	initiatingMessage.Value.DownlinkNASTransport = new(ngapType.DownlinkNASTransport)

	downlinkNasTransport := initiatingMessage.Value.DownlinkNASTransport
	downlinkNasTransportIEs := &downlinkNasTransport.ProtocolIEs

	// AMF UE NGAP ID
	ie := ngapType.DownlinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.DownlinkNASTransportIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = ue.GetAmfNgapId()

	downlinkNasTransportIEs.List = append(downlinkNasTransportIEs.List, ie)

	// RAN UE NGAP ID
	ie = ngapType.DownlinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.DownlinkNASTransportIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ue.GetRanNgapId()

	downlinkNasTransportIEs.List = append(downlinkNasTransportIEs.List, ie)

	// NAS PDU
	ie = ngapType.DownlinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDNASPDU
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.DownlinkNASTransportIEsPresentNASPDU
	ie.Value.NASPDU = new(ngapType.NASPDU)

	ie.Value.NASPDU.Value = nasPdu

	downlinkNasTransportIEs.List = append(downlinkNasTransportIEs.List, ie)

	return
}
