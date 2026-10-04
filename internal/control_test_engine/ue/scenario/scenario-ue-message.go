/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package scenario

import "my5G-RANTester/internal/control_test_engine/procedures"

type ScenarioMessage struct {
	StateChange int
	// Control is set, with the UE's Status, when the UE answers a control request
	// rather than changing state.
	Control *procedures.ControlRequest
	Status  procedures.UeStatus
}
