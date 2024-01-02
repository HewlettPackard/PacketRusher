/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

func TestAttachUeWithConfiguration(tunnelEnabled bool) {
	TestMultiUesInQueue(1, tunnelEnabled, true, false, 500, 0, 0, 0, 0, 1)
}
