/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Valentin D'Emmanuele
 */
package mm_5gs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

func ServiceRequest(ue *context.UEContext) (nasPdu []byte) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeServiceRequest)

	serviceRequest := nasMessage.NewServiceRequest(0)
	serviceRequest.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	serviceRequest.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	serviceRequest.SetMessageType(nas.MsgTypeServiceRequest)
	serviceRequest.SetServiceTypeValue(0x01)
	serviceRequest.SetNasKeySetIdentifiler(0x01)
	serviceRequest.SetAMFSetID(ue.GetAmfSetId())
	serviceRequest.SetAMFPointer(ue.GetAmfPointer())
	serviceRequest.SetTypeOfIdentity(4) // 5G-S-TMSI
	serviceRequest.SetTMSI5G(ue.GetTMSI5G())
	serviceRequest.TMSI5GS.SetLen(7)

	serviceRequest.UplinkDataStatus = new(nasType.UplinkDataStatus)
	serviceRequest.UplinkDataStatus.SetIei(nasMessage.ServiceRequestUplinkDataStatusType)
	serviceRequest.UplinkDataStatus.SetLen(2)

	pduFlag := uint16(0)
	for i, pduSession := range ue.PduSession {
		pduFlag = pduFlag + (boolToUint16(pduSession != nil) << (i + 1))
	}

	serviceRequest.UplinkDataStatus.Buffer = make([]byte, 2)
	binary.LittleEndian.PutUint16(serviceRequest.UplinkDataStatus.Buffer, pduFlag)

	serviceRequest.PDUSessionStatus = new(nasType.PDUSessionStatus)
	serviceRequest.PDUSessionStatus.SetIei(nasMessage.ServiceRequestPDUSessionStatusType)
	serviceRequest.PDUSessionStatus.SetLen(2)
	serviceRequest.PDUSessionStatus.Buffer = serviceRequest.UplinkDataStatus.Buffer

	m.GmmMessage.ServiceRequest = serviceRequest

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()

	serviceRequest.UplinkDataStatus = nil
	serviceRequest.PDUSessionStatus = nil

	serviceRequest.NASMessageContainer = nasType.NewNASMessageContainer(nasMessage.ServiceRequestNASMessageContainerType)
	serviceRequest.NASMessageContainer.SetLen(uint16(len(nasPdu)))
	serviceRequest.NASMessageContainer.Buffer = nasPdu

	data = new(bytes.Buffer)
	err = m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()

	return
}

func boolToUint16(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}
