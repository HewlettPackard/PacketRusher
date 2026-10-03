/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"fmt"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
	"strings"
)

func getSecurityModeComplete(container []uint8, digits string) []byte {
	identity := &ie.MobileId5GS{TypeOfId: ie.IdType_5GS_IMEISV}
	for i := range identity.IMEISV {
		identity.IMEISV[i] = digits[i] - '0'
	}
	msg := &nas.SecModeComplete{IMEISV: identity}
	if container != nil {
		msg.NASMsgCntr = &ie.NASMsgCntr{Contents: container}
	}
	return encodePlain(msg)
}
func SecurityModeComplete(ue *context.UEContext, rinmr uint8) ([]byte, error) {
	// Keep retransmitting the registration request for cores that do not set RINMR.
	registration := GetRegistrationRequest(ie.RegType_InitialReg, nil, nil, true, ue)
	pdu := getSecurityModeComplete(registration, imeisvFromMsin(ue.GetMsin()))
	pdu, err := nas_control.EncodeNasPduWithSecurity(ue, pdu, nas.SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx, true, true)
	if err != nil {
		return nil, fmt.Errorf("encoding Security Mode Complete: %w", err)
	}
	return pdu, nil
}
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
