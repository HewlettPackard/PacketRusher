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

// TS 24.501 8.2.26
func getSecurityModeComplete(nasMessageContainer []uint8) (nasPdu []byte) {

	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeSecurityModeComplete)

	securityModeComplete := nasMessage.NewSecurityModeComplete(0)
	securityModeComplete.ExtendedProtocolDiscriminator.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	// TODO: modify security header type if need security protected
	securityModeComplete.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	securityModeComplete.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	securityModeComplete.SecurityModeCompleteMessageIdentity.SetMessageType(nas.MsgTypeSecurityModeComplete)

	securityModeComplete.IMEISV = nasType.NewIMEISV(nasMessage.SecurityModeCompleteIMEISVType)
	securityModeComplete.IMEISV.SetLen(9)
	securityModeComplete.SetOddEvenIdic(0)
	securityModeComplete.SetTypeOfIdentity(nasMessage.MobileIdentity5GSTypeImeisv)
	securityModeComplete.SetIdentityDigit1(1)
	securityModeComplete.SetIdentityDigitP_1(1)
	securityModeComplete.SetIdentityDigitP(1)

	if nasMessageContainer != nil {
		securityModeComplete.NASMessageContainer = nasType.NewNASMessageContainer(nasMessage.SecurityModeCompleteNASMessageContainerType)
		securityModeComplete.NASMessageContainer.SetLen(uint16(len(nasMessageContainer)))
		securityModeComplete.NASMessageContainer.SetNASMessageContainerContents(nasMessageContainer)
	}

	m.GmmMessage.SecurityModeComplete = securityModeComplete

	data := new(bytes.Buffer)
	err := m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	nasPdu = data.Bytes()
	return
}

func SecurityModeComplete(ue *context.UEContext, rinmr uint8) ([]byte, error) {
	var registrationRequest []byte

	// ueSecurityCapability := context.SetUESecurityCapability(ue)

	/*
		requestedNssai := new(nasType.RequestedNSSAI)
		nssai := nasConvert.SnssaiToNas(models.Snssai{Sst: ue.Snssai.Sst, Sd: ue.Snssai.Sd})
		requestedNssai.Buffer = nssai
		requestedNssai.Len = uint8(len(nssai))
		requestedNssai.Iei = nasMessage.RegistrationRequestRequestedNSSAIType
	*/
	if rinmr == 1 {
		registrationRequest = GetRegistrationRequest(nasMessage.RegistrationType5GSInitialRegistration, nil, nil, true, ue)
	} else {
		// TODO: free5gc does not send rinmr and wait for restransmission of registration request
		// registrationRequest = nil
		registrationRequest = GetRegistrationRequest(nasMessage.RegistrationType5GSInitialRegistration, nil, nil, true, ue)
	}

	pdu := getSecurityModeComplete(registrationRequest)
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE  NAS Security Mode Complete message", ue.UeSecurity.Supi)
	}
	return pdu, nil
}
