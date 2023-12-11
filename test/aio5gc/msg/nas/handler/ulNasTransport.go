/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package handler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg"
	"slices"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/openapi/models"
)

func UlNasTransport(nasReq *nas.Message, gnb *context.GNBContext, ue *context.UEContext, session *context.SessionContext) error {

	ulNasTransport := nasReq.ULNASTransport
	var err error
	if !ue.GetInitialContextSetup() {
		return errors.New("[5GC][NGAP] This UE has no security context set up")
	}

	switch ulNasTransport.GetPayloadContainerType() {
	// TS 24.501 5.4.5.2.3 case a)
	case nasMessage.PayloadContainerTypeN1SMInfo:
		err = transport5GSMMessage(ue, ulNasTransport, session, gnb)

	default:
		err = fmt.Errorf("[5GC][NAS] Payload Container type not implemented")
	}

	if err != nil {
		return err
	}
	return nil
}

func transport5GSMMessage(ue *context.UEContext, ulNasTransport *nasMessage.ULNASTransport, session *context.SessionContext, gnb *context.GNBContext) error {
	requestType := ulNasTransport.RequestType
	n1smContent := ulNasTransport.PayloadContainer.GetPayloadContainerContents()
	var pduSessionID int32

	if id := ulNasTransport.PduSessionID2Value; id != nil {
		pduSessionID = int32(id.GetPduSessionID2Value())
	} else {
		return errors.New("[5GC][NAS] PDU Session ID is nil")
	}

	if requestType == nil {
		n1smContent := ulNasTransport.PayloadContainer.GetPayloadContainerContents()
		return handleUnspecifiedRequest(n1smContent, ue, pduSessionID, gnb)
	}

	var (
		snssai models.Snssai
		dnn    string
	)
	// If the S-NSSAI IE is not included, select a default snssai
	if ulNasTransport.SNSSAI != nil {
		snssai = nasConvert.SnssaiToModels(ulNasTransport.SNSSAI)
	} else {
		snssai = ue.GetNssai()
	}

	dnnList := session.GetDnnList()
	if ulNasTransport.DNN != nil {
		if !slices.Contains(dnnList, ulNasTransport.DNN.GetDNN()) {
			return errors.New("[5GC] Unknown DNN requested")
		}
		dnn = ulNasTransport.DNN.GetDNN()

	} else {
		dnn = dnnList[0]
	}

	switch requestType.GetRequestTypeValue() {
	// case iii) if the AMF does not have a PDU session routing context for the PDU session ID and the UE
	// and the Request type IE is included and is set to "initial request"
	case nasMessage.ULNASTransportRequestTypeInitialRequest:
		return handleInitialRequest(n1smContent, ue, session, pduSessionID, snssai, dnn, gnb)

	default:
		return errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
}

func handleUnspecifiedRequest(n1smContent []uint8,
	ue *context.UEContext,
	pduSessionID int32,
	gnb *context.GNBContext) error {

	m := nas.NewMessage()
	err := m.GsmMessageDecode(&n1smContent)
	if err != nil {
		return errors.New("[5GC][NAS] GsmMessageDecode Error: " + err.Error())
	}
	switch m.GsmHeader.GetMessageType() {
	case nas.MsgTypePDUSessionReleaseRequest:
		smContext, err := context.ReleasePDUSession(ue, pduSessionID)
		if err != nil {
			return err
		}
		msg.SendPDUSessionReleaseCommand(gnb, ue, smContext, nasMessage.Cause5GSMRegularDeactivation)

	default:
		return errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
	return nil
}

func handleInitialRequest(n1smContent []uint8,
	ue *context.UEContext,
	session *context.SessionContext,
	pduSessionID int32,
	snssai models.Snssai,
	dnn string,
	gnb *context.GNBContext) error {

	m := nas.NewMessage()
	err := m.GsmMessageDecode(&n1smContent)
	if err != nil {
		return errors.New("[5GC][NAS] GsmMessageDecode Error: " + err.Error())
	}
	if m.GsmHeader.GetMessageType() != nas.MsgTypePDUSessionEstablishmentRequest {
		return errors.New("[5GC][NAS] UL NAS Transport container message expected to be PDU Session Establishment Request but was not")
	}
	sessionRequest := m.PDUSessionEstablishmentRequest

	smContext, err := context.CreatePDUSession(sessionRequest, ue, session, pduSessionID, snssai, dnn)
	if err != nil {
		return err
	}
	msg.SendPDUSessionEstablishmentAccept(gnb, ue, smContext, session)
	return nil
}
