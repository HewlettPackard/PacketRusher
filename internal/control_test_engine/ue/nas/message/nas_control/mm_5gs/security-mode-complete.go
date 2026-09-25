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
	"strings"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

// TS 24.501 8.2.26
func getSecurityModeComplete(nasMessageContainer []uint8, imeisv string) (nasPdu []byte) {

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
	setImeisvDigits(securityModeComplete.IMEISV, imeisv)

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

	pdu := getSecurityModeComplete(registrationRequest, imeisvFromMsin(ue.GetMsin()))
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		return nil, fmt.Errorf("Error encoding %s IMSI UE  NAS Security Mode Complete message", ue.UeSecurity.Supi)
	}
	return pdu, nil
}

// imeisvFromMsin derives a 16-digit IMEISV from the UE's MSIN, so that every simulated UE
// reports a distinct PEI: the MSIN, left-padded with zeros, fills the 14 TAC and SNR digits,
// followed by software version 01.
func imeisvFromMsin(msin string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msin)
	if len(digits) > 14 {
		digits = digits[len(digits)-14:]
	}
	return strings.Repeat("0", 14-len(digits)) + digits + "01"
}

// setImeisvDigits BCD-encodes a 16-digit IMEISV after the octet carrying digit 1, the odd/even
// indicator and the type of identity. digits must be exactly 16 decimal digits, as
// imeisvFromMsin returns. TS 24.501 9.11.3.4: bits 5 to 8 of the last octet are
// filled with an end mark coded as "1111". Without it the digit count is 17 and a strict
// decoder rejects the PEI.
func setImeisvDigits(imeisv *nasType.IMEISV, digits string) {
	imeisv.SetIdentityDigit1(digits[0] - '0')
	for i := 1; i < 9; i++ {
		low := digits[2*i-1] - '0'
		high := uint8(0x0f)
		if 2*i < len(digits) {
			high = digits[2*i] - '0'
		}
		imeisv.Octet[i] = high<<4 | low
	}
}
