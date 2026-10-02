/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func IdentityResponse(ue *context.UEContext) []byte {
	identity := ue.GetSuci()
	return encodePlain(&nas.IdRsp{MobileId: &identity})
}
