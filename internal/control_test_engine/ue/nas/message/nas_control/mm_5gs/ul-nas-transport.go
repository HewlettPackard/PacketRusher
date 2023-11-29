/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/sm_5gs"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"

	"github.com/free5gc/openapi/models"
)

func Request_UlNasTransport(pduSession *context.UEPDUSession, ue *context.UEContext) ([]byte, error) {

	pdu := getUlNasTransport_PduSessionEstablishmentRequest(pduSession.Id, ue.Dnn, &ue.Snssai)
	if pdu == nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}

	return pdu, nil
}

func Release_UlNasTransport(pduSession *context.UEPDUSession, ue *context.UEContext) ([]byte, error) {

	pdu := getUlNasTransport_PduSessionEstablishmentRelease(pduSession.Id)
	if pdu == nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}

	return pdu, nil
}

func getUlNasTransport_PduSessionEstablishmentRequest(pduSessionId uint8, dnn string, sNssai *models.Snssai) (nasPdu []byte) {

	pduSessionEstablishmentRequest := sm_5gs.GetPduSessionEstablishmentRequest(pduSessionId)

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeULNASTransport)

	ulNasTransport := nasMessage.NewULNASTransport(0)
	ulNasTransport.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	ulNasTransport.SetMessageType(nas.MsgTypeULNASTransport)
	ulNasTransport.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	ulNasTransport.PduSessionID2Value = new(nasType.PduSessionID2Value)
	ulNasTransport.PduSessionID2Value.SetIei(nasMessage.ULNASTransportPduSessionID2ValueType)
	ulNasTransport.PduSessionID2Value.SetPduSessionID2Value(pduSessionId)
	ulNasTransport.RequestType = new(nasType.RequestType)
	ulNasTransport.RequestType.SetIei(nasMessage.ULNASTransportRequestTypeType)
	ulNasTransport.RequestType.SetRequestTypeValue(nasMessage.ULNASTransportRequestTypeInitialRequest)

	if dnn != "" {
		ulNasTransport.DNN = new(nasType.DNN)
		ulNasTransport.DNN.SetIei(nasMessage.ULNASTransportDNNType)
		ulNasTransport.DNN.SetLen(uint8(len(dnn)))
		ulNasTransport.DNN.SetDNN(dnn)
	}

	if sNssai != nil {
		ulNasTransport.SNSSAI = nasType.NewSNSSAI(nasMessage.ULNASTransportSNSSAIType)
		if sNssai.Sd == "" {
			ulNasTransport.SNSSAI.SetLen(1)
		} else {
			ulNasTransport.SNSSAI.SetLen(4)
			var sdTemp [3]uint8
			sd, _ := hex.DecodeString(sNssai.Sd)
			copy(sdTemp[:], sd)
			ulNasTransport.SNSSAI.SetSD(sdTemp)
		}
		ulNasTransport.SNSSAI.SetSST(uint8(sNssai.Sst))
	}

	ulNasTransport.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(nasMessage.PayloadContainerTypeN1SMInfo)
	ulNasTransport.PayloadContainer.SetLen(uint16(len(pduSessionEstablishmentRequest)))
	ulNasTransport.PayloadContainer.SetPayloadContainerContents(pduSessionEstablishmentRequest)

	m.GmmMessage.ULNASTransport = ulNasTransport

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}

func getUlNasTransport_PduSessionEstablishmentRelease(pduSessionId uint8) (nasPdu []byte) {

	pduSessionReleaseRequest := sm_5gs.GetPduSessionReleaseRequest(pduSessionId)

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeULNASTransport)

	ulNasTransport := nasMessage.NewULNASTransport(0)
	ulNasTransport.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	ulNasTransport.SetMessageType(nas.MsgTypeULNASTransport)
	ulNasTransport.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	ulNasTransport.PduSessionID2Value = new(nasType.PduSessionID2Value)
	ulNasTransport.PduSessionID2Value.SetIei(nasMessage.ULNASTransportPduSessionID2ValueType)
	ulNasTransport.PduSessionID2Value.SetPduSessionID2Value(pduSessionId)

	ulNasTransport.SpareHalfOctetAndPayloadContainerType.SetPayloadContainerType(nasMessage.PayloadContainerTypeN1SMInfo)
	ulNasTransport.PayloadContainer.SetLen(uint16(len(pduSessionReleaseRequest)))
	ulNasTransport.PayloadContainer.SetPayloadContainerContents(pduSessionReleaseRequest)

	m.GmmMessage.ULNASTransport = ulNasTransport

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
