/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"sync"
)

func TestAttachGnbWithConfiguration() {

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// wrong messages:
	// cfg.GNodeB.PlmnList.Mcc = "891"
	// cfg.GNodeB.PlmnList.Mnc = "23"
	// cfg.GNodeB.PlmnList.Tac = "000002"
	// cfg.GNodeB.SliceSupportList.St = "10"
	// cfg.GNodeB.SliceSupportList.Sst = "010239"

	go gnb.InitGnb(cfg.GNodeB, cfg.AMFs, &wg)

	wg.Add(1)

	wg.Wait()
}
