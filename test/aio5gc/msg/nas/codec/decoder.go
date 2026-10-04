/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package codec

import (
	"fmt"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
)

func DecodePlainNasNoIntegrityCheck(payload []byte) (nas.Message, error) {
	// A 5GSM header carries a session ID in octet 2, not a security header.
	if len(payload) > 0 && payload[0] == byte(nas.Epd5GSSessMgmtMsg) {
		return nas.Parse(payload, nil)
	}
	st := nas.GetSecHdrType(payload)
	if st == nas.SecHdrTypeIntegrityProtectedAndCiphered || st == nas.SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx {
		return nil, fmt.Errorf("NAS payload is ciphered")
	}
	if st != nas.SecHdrTypePlainNas {
		if len(payload) < int(nas.SecHdrLen) {
			return nil, fmt.Errorf("truncated NAS security header")
		}
		payload = payload[nas.SecHdrLen:]
	}
	return nas.Parse(payload, nil)
}
func Decode(ue *context.UEContext, payload []byte, initialMessage bool) (nas.Message, bool, error) {
	if ue == nil || ue.GetSecurityContext() == nil {
		return nil, false, fmt.Errorf("NAS security context is nil")
	}
	sc := ue.GetSecurityContext().NASSecurityContext()
	candidate := sc.Clone()
	msg, err := nas.Parse(payload, candidate)
	if err != nil {
		return nil, false, err
	}
	ue.GetSecurityContext().SetULCount(*candidate.UplinkCount)
	ue.GetSecurityContext().SetDLCount(*candidate.DownlinkCount)
	return msg, payload[0] == byte(nas.Epd5GSMobilityMgmtMsg) && nas.GetSecHdrType(payload) != nas.SecHdrTypePlainNas, nil
}
