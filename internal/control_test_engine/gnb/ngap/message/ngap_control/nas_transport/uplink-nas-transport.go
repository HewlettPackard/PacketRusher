/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas_transport

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap"

	"github.com/free5gc/ngap/ngapType"
)

func getUplinkNASTransport(amfUeNgapID, ranUeNgapID int64, nasPdu []byte, gnb *context.GNBContext) ([]byte, error) {
	message := buildUplinkNasTransport(amfUeNgapID, ranUeNgapID, nasPdu, gnb)
	return ngap.Encoder(message)
}

func buildUplinkNasTransport(amfUeNgapID, ranUeNgapID int64, nasPdu []byte, gnb *context.GNBContext) (pdu ngapType.NGAPPDU) {

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeUplinkNASTransport
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentIgnore

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentUplinkNASTransport
	initiatingMessage.Value.UplinkNASTransport = new(ngapType.UplinkNASTransport)

	uplinkNasTransport := initiatingMessage.Value.UplinkNASTransport
	uplinkNasTransportIEs := &uplinkNasTransport.ProtocolIEs

	// AMF UE NGAP ID
	ie := ngapType.UplinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.UplinkNASTransportIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	uplinkNasTransportIEs.List = append(uplinkNasTransportIEs.List, ie)

	// RAN UE NGAP ID
	ie = ngapType.UplinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.UplinkNASTransportIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	uplinkNasTransportIEs.List = append(uplinkNasTransportIEs.List, ie)

	// NAS-PDU
	ie = ngapType.UplinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDNASPDU
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.UplinkNASTransportIEsPresentNASPDU
	ie.Value.NASPDU = new(ngapType.NASPDU)

	// TODO: complete NAS-PDU
	nASPDU := ie.Value.NASPDU
	nASPDU.Value = nasPdu

	uplinkNasTransportIEs.List = append(uplinkNasTransportIEs.List, ie)

	// User Location Information
	ie = ngapType.UplinkNASTransportIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUserLocationInformation
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.UplinkNASTransportIEsPresentUserLocationInformation
	ie.Value.UserLocationInformation = new(ngapType.UserLocationInformation)

	userLocationInformation := ie.Value.UserLocationInformation
	userLocationInformation.Present = ngapType.UserLocationInformationPresentUserLocationInformationNR
	userLocationInformation.UserLocationInformationNR = new(ngapType.UserLocationInformationNR)

	userLocationInformationNR := userLocationInformation.UserLocationInformationNR
	userLocationInformationNR.NRCGI.PLMNIdentity = gnb.GetPLMNIdentity()
	userLocationInformationNR.NRCGI.NRCellIdentity = gnb.GetNRCellIdentity()

	userLocationInformationNR.TAI.PLMNIdentity = gnb.GetPLMNIdentity()
	userLocationInformationNR.TAI.TAC.Value = gnb.GetTacInBytes()

	uplinkNasTransportIEs.List = append(uplinkNasTransportIEs.List, ie)

	return
}

func SendUplinkNasTransport(message []byte, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {

	sendMsg, err := getUplinkNASTransport(ue.GetAmfUeId(), ue.GetRanUeId(), message, gnb)
	if err != nil {
		return nil, fmt.Errorf("Error getting UE Id %d NAS Authentication Response", ue.GetRanUeId())
	}

	return sendMsg, nil
}

/*
func UplinkNasTransport(connN2 *sctp.SCTPConn, amfUeNgapID int64, ranUeNgapID int64, nasPdu []byte, gnb *context.RanGnbContext) error {

	sendMsg, err := getUplinkNASTransport(amfUeNgapID, ranUeNgapID, nasPdu, gnb.GetMccAndMncInOctets(), gnb.GetTacInBytes())
	if err != nil {
		return fmt.Errorf("Error getting ueId %d NAS Authentication Response", ranUeNgapID)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending ueId %d NAS Authentication Response", ranUeNgapID)
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": "AMF",
		"message":     "UPLINK NAS TRANSPORT",
	}).Info("Sending message")

	return nil
}
*/
