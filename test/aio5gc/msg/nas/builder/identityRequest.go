/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023-2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func IdentityRequest(ue *context.UEContext) ([]byte, error) {
	st := nas.SecHdrTypePlainNas
	if ue.GetState().Is(context.Registered) {
		st = nas.SecHdrTypeIntegrityProtectedAndCiphered
	}
	return codec.Encode(ue, &nas.IdReq{IdType: &ie.IdType5GS{IdType: ie.IdType_5GS_SUCI}}, st)
}
