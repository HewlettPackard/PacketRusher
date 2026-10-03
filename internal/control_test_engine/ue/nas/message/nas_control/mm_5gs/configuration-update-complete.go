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

func ConfigurationUpdateComplete(ue *context.UEContext) ([]byte, error) {
	return nas_control.EncodeNasPduWithSecurity(ue, getConfigurationUpdateComplete(), nas.SecHdrTypeIntegrityProtectedAndCiphered, true, false)
}
func getConfigurationUpdateComplete() []byte { return encodePlain(&nas.CfgUpdateComplete{}) }
