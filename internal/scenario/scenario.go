/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"my5G-RANTester/config"
)

// Description of the scenario to be run for one UE
type UEScenario struct {
	Config  config.Ue
	Tasks   []Task
	Loop    bool // Restart scenario once done
	Persist bool // Keep scenario going after all tasks are done
}
