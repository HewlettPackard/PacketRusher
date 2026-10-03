/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
)

func RegistrationComplete(ue *context.UEContext) ([]byte, error) {
	return nas_control.EncodeNasPduWithSecurity(ue, getRegistrationComplete(nil), nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
}
func getRegistrationComplete(sorTransparentContainer []uint8) []byte {
	// Upstream has no SORTransparentCntr codec yet. No caller requests SOR data.
	if sorTransparentContainer != nil {
		return nil
	}
	return encodePlain(&nas.RegComplete{})
}
