/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package templates

import "my5G-RANTester/config"

func TestAttachUeWithConfiguration(tunnelEnabled bool) {
	tunnelMode := config.TunnelDisabled
	if tunnelEnabled {
		tunnelMode = config.TunnelVrf
	}
	TestMultiUesInQueue(1, tunnelMode, true, false, 0, 200, 500, 0, 0, 0, 0, 0, 1, 0, "")
}
