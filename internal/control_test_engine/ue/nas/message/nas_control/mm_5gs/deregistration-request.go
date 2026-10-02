/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package mm_5gs

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control"
)

func DeregistrationRequest(ue *context.UEContext) ([]byte, error) {
	return nas_control.EncodeNasPduWithSecurity(ue, getDeregistrationRequest(ue), nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
}
func getDeregistrationRequest(ue *context.UEContext) []byte {
	identity := ue.GetSuci()
	if ue.Get5gGuti() != nil {
		identity = *ue.Get5gGuti()
	}
	return encodePlain(&nas.DeregReqUEOrig{
		DeregType: &ie.DeregType{Switchoff: true, AccessType: ie.AccessType_3gpp},
		Ngksi:     &ue.UeSecurity.NgKsi, MobileId5GS: &identity,
	})
}
