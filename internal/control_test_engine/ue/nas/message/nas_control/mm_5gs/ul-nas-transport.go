/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"fmt"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/openapi/models"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/sm_5gs"
)

func Request_UlNasTransport(pduSession *context.UEPDUSession, ue *context.UEContext) ([]byte, error) {

	pdu := getUlNasTransport_PduSessionEstablishmentRequest(pduSession.Id, ue.Dnn, &ue.Snssai)
	if pdu == nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
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
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}

	return pdu, nil
}

func ReleasComplete_UlNasTransport(pduSession *context.UEPDUSession, ue *context.UEContext) ([]byte, error) {

	pdu := getUlNasTransport_PduSessionReleaseComplete(pduSession.Id, ue.Dnn, &ue.Snssai)
	if pdu == nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE PduSession Establishment Request Msg", ue.UeSecurity.Supi)
	}

	return pdu, nil
}

func transport(id uint8, payload []byte, request *ie.ReqType, dnn string, snssai *models.Snssai) []byte {
	msg := &nas.ULNASTransport{
		PayloadCntrType: &ie.PayloadCntrType{Value: ie.PayloadCntrType_N1SMInfo},
		PayloadCntr:     &ie.PayloadCntr{Pct: ie.PayloadCntrType_N1SMInfo, Contents: payload},
		PDUSessID:       &ie.PDUSessId2{Value: id}, ReqType: request,
	}
	if dnn != "" {
		msg.DNN = &ie.DNN{Value: dnn}
	}
	if snssai != nil {
		msg.SNSSAI = &ie.SNSSAI{SST: uint8(snssai.Sst), SD: snssai.Sd}
	}
	return encodePlain(msg)
}
func getUlNasTransport_PduSessionEstablishmentRequest(id uint8, dnn string, snssai *models.Snssai) []byte {
	return transport(id, sm_5gs.GetPduSessionEstablishmentRequest(id), &ie.ReqType{Value: ie.ReqType_InitialReq}, dnn, snssai)
}
func getUlNasTransport_PduSessionEstablishmentRelease(id uint8) []byte {
	return transport(id, sm_5gs.GetPduSessionReleaseRequest(id), nil, "", nil)
}
func getUlNasTransport_PduSessionReleaseComplete(id uint8, dnn string, snssai *models.Snssai) []byte {
	return transport(id, sm_5gs.GetPduSessionReleaseComplete(id), &ie.ReqType{Value: ie.ReqType_ExistingPDUSess}, dnn, snssai)
}
