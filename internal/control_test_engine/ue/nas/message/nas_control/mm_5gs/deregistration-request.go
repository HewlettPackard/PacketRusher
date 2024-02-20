/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"bytes"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

func DeregistrationRequest(ue *context.UEContext) ([]byte, error) {

	pdu := getDeregistrationRequest(ue)
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		return nil, fmt.Errorf("error encoding %s IMSI UE  NAS Security Mode Complete message", ue.UeSecurity.Supi)
	}
	return pdu, nil
}

func getDeregistrationRequest(ue *context.UEContext) (nasPdu []byte) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration)

	deregistrationRequest := nasMessage.NewDeregistrationRequestUEOriginatingDeregistration(0)

	deregistrationRequest.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	deregistrationRequest.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	deregistrationRequest.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0x00)
	deregistrationRequest.SetSwitchOff(1)
	deregistrationRequest.SetReRegistrationRequired(0)
	deregistrationRequest.SetAccessType(1)
	deregistrationRequest.DeregistrationRequestMessageIdentity.SetMessageType(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration)
	deregistrationRequest.NgksiAndDeregistrationType.SetTSC(nasMessage.TypeOfSecurityContextFlagNative)

	deregistrationRequest.NgksiAndDeregistrationType.SetNasKeySetIdentifiler(uint8(ue.UeSecurity.NgKsi.Ksi))
	// If AMF previously assigned the UE a 5G-GUTI, reuses it
	// If the 5G-GUTI is no longer valid, AMF will issue an Identity Request
	// which we'll answer with the requested Mobility Identity (eg. SUCI)
	if ue.Get5gGuti() != nil {
		guti := ue.Get5gGuti()
		deregistrationRequest.MobileIdentity5GS = nasType.MobileIdentity5GS{
			Iei:    guti.Iei,
			Len:    guti.Len,
			Buffer: guti.Octet[:],
		}
	} else {
		deregistrationRequest.MobileIdentity5GS = ue.GetSuci()
	}
	m.GmmMessage.DeregistrationRequestUEOriginatingDeregistration = deregistrationRequest

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}
