/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func PDUSessionReleaseCommand(ue *context.UEContext, sm context.SmContext, cause uint8) ([]byte, error) {
	b, err := buildPDUSessionReleaseCommand(ue, sm, cause).MarshalBinary()
	if err != nil {
		return nil, err
	}
	msg, err := buildDLNASTransport(ue, b, uint8(sm.GetPduSessionId()))
	if err != nil {
		return nil, err
	}
	return codec.Encode(ue, msg, nas.SecHdrTypeIntegrityProtectedAndCiphered)
}
func buildPDUSessionReleaseCommand(ue *context.UEContext, sm context.SmContext, cause uint8) *nas.PDUSessRelCmd {
	return &nas.PDUSessRelCmd{PDUSessId: uint8(sm.GetPduSessionId()), PTI: sm.GetPti(), Cause5GSM: &ie.Cause5GSM{Value: cause}}
}
