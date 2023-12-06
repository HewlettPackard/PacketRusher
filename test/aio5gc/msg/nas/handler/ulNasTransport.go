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
	smMessage := ulNasTransport.PayloadContainer.GetPayloadContainerContents()
	var pduSessionID int32

	if requestType == nil {
		return errors.New("[5GC][NAS] ulNasTransport Request type is nil")
	}

	if id := ulNasTransport.PduSessionID2Value; id != nil {
		pduSessionID = int32(id.GetPduSessionID2Value())
	} else {
		return errors.New("[5GC][NAS] PDU Session ID is nil")
	}

	switch requestType.GetRequestTypeValue() {
	// case iii) if the AMF does not have a PDU session routing context for the PDU session ID and the UE
	// and the Request type IE is included and is set to "initial request"
	case nasMessage.ULNASTransportRequestTypeInitialRequest:
		return ulNASTransportRequestTypeInitialRequest(ulNasTransport, ue, session, pduSessionID, smMessage, gnb)
	default:
		return errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
}

func ulNASTransportRequestTypeInitialRequest(initialRequest *nasMessage.ULNASTransport,
	ue *context.UEContext,
	session *context.SessionContext,
	pduSessionID int32,
	smMessage []uint8,
	gnb *context.GNBContext) error {

	var (
		snssai models.Snssai
		dnn    string
	)
	// If the S-NSSAI IE is not included, select a default snssai
	if initialRequest.SNSSAI != nil {
		snssai = nasConvert.SnssaiToModels(initialRequest.SNSSAI)
	} else {
		snssai = ue.GetNssai()
	}

	dnnList := session.GetDnnList()
	if initialRequest.DNN != nil {
		if !slices.Contains(dnnList, initialRequest.DNN.GetDNN()) {
			return errors.New("[5GC] Unknown DNN requested")
		}
		dnn = initialRequest.DNN.GetDNN()

	} else {
		dnn = dnnList[0]
	}

	n1smContent := initialRequest.PayloadContainer.GetPayloadContainerContents()
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
