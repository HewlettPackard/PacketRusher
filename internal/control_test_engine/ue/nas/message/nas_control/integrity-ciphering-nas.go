/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas_control

import (
	"fmt"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func EncodeNasPduWithSecurity(ue *context.UEContext, pdu []byte, securityHeaderType nas.SecHdrType, securityContextAvailable, newSecurityContext bool) ([]byte, error) {
	if ue == nil {
		return nil, fmt.Errorf("UE context is nil")
	}
	msg, err := nas.Parse(pdu, nil)
	if err != nil {
		return nil, err
	}
	if !securityContextAvailable {
		return msg.MarshalBinary()
	}
	if newSecurityContext {
		ue.UeSecurity.ULCount.Set(0, 0)
		ue.UeSecurity.DLCount.Set(0, 0)
	}
	return nas.Marshal(msg, ue.NASSecurityContext(), securityHeaderType)
}
