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

	ie "github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/openapi/models"
)

func UlNasTransport(nasReq *nas.ULNASTransport, gnb *context.GNBContext, ue *context.UEContext, session *context.SessionContext) error {
	var err error
	switch ue.GetState().Current() {
	case context.Registered:
		return DefaultUlNasTransport(nasReq, gnb, ue, session)
	default:
		err = fmt.Errorf("[5GC][NAS] Unexpected message: received %s for UlNasTransport", ue.GetState().Current())
	}
	return err
}

func DefaultUlNasTransport(nasReq *nas.ULNASTransport, gnb *context.GNBContext, ue *context.UEContext, session *context.SessionContext) error {

	ulNasTransport := nasReq
	var err error

	switch ulNasTransport.PayloadCntrType.Value {
	// TS 24.501 5.4.5.2.3 case a)
	case ie.PayloadCntrType_N1SMInfo:
		err = transport5GSMMessage(ue, ulNasTransport, session, gnb)

	default:
		err = fmt.Errorf("[5GC][NAS] Payload Container type not implemented")
	}

	if err != nil {
		return err
	}
	return nil
}

func transport5GSMMessage(ue *context.UEContext, ulNasTransport *nas.ULNASTransport, session *context.SessionContext, gnb *context.GNBContext) error {
	requestType := ulNasTransport.ReqType
	n1smContent := ulNasTransport.PayloadCntr.Contents
	var pduSessionID int32

	if id := ulNasTransport.PDUSessID; id != nil {
		pduSessionID = int32(id.Value)
	} else {
		return errors.New("[5GC][NAS] PDU Session ID is nil")
	}

	if requestType == nil {
		n1smContent := ulNasTransport.PayloadCntr.Contents
		return handleUnspecifiedRequest(n1smContent, ue, pduSessionID, gnb)
	}

	var (
		snssai models.Snssai
		dnn    string
	)
	// If the S-NSSAI IE is not included, select a default snssai
	if ulNasTransport.SNSSAI != nil {
		snssai = models.Snssai{Sst: int32(ulNasTransport.SNSSAI.SST), Sd: ulNasTransport.SNSSAI.SD}
	} else {
		snssai = ue.GetDefaultSNssai()
	}

	dnnList := session.GetDnnList()
	if ulNasTransport.DNN != nil {
		if !slices.Contains(dnnList, ulNasTransport.DNN.Value) {
			return errors.New("[5GC] Unknown DNN requested")
		}
		dnn = ulNasTransport.DNN.Value

	} else {
		dnn = dnnList[0]
	}

	switch requestType.Value {
	// case iii) if the AMF does not have a PDU session routing context for the PDU session ID and the UE
	// and the Request type IE is included and is set to "initial request"
	case ie.ReqType_InitialReq:
		return handleInitialRequest(n1smContent, ue, session, pduSessionID, snssai, dnn, gnb)

	default:
		return errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
}

func handleUnspecifiedRequest(n1smContent []uint8,
	ue *context.UEContext,
	pduSessionID int32,
	gnb *context.GNBContext) error {

	m, err := nas.Parse(n1smContent, nil)
	if err != nil {
		return errors.New("[5GC][NAS] GsmMessageDecode Error: " + err.Error())
	}
	switch m.MsgType() {
	case nas.MsgTypePDUSessRelReq:
		smContext, err := context.ReleasePDUSession(ue, pduSessionID)
		if err != nil {
			return err
		}
		return msg.SendPDUSessionReleaseCommand(gnb, ue, smContext, ie.Cause5GSM_RegularDeactivation)

	default:
		return errors.New("[5GC][NAS] Unimplemented ulNasTransport Request type")
	}
}

func handleInitialRequest(n1smContent []uint8,
	ue *context.UEContext,
	session *context.SessionContext,
	pduSessionID int32,
	snssai models.Snssai,
	dnn string,
	gnb *context.GNBContext) error {

	m, err := nas.Parse(n1smContent, nil)
	if err != nil {
		return errors.New("[5GC][NAS] GsmMessageDecode Error: " + err.Error())
	}
	if m.MsgType() != nas.MsgTypePDUSessEstReq {
		return errors.New("[5GC][NAS] UL NAS Transport container message expected to be PDU Session Establishment Request but was not")
	}
	sessionRequest := m.(*nas.PDUSessEstReq)

	smContext, err := context.CreatePDUSession(sessionRequest, ue, session, pduSessionID, snssai, dnn)
	if err != nil {
		return err
	}
	return msg.SendPDUSessionEstablishmentAccept(gnb, ue, smContext, session)
}
