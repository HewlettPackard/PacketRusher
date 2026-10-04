/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package templates

import (
	"context"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"os"
	"os/signal"
	"sync"
)

func TestAttachGnbWithConfiguration() error {

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// wrong messages:
	// cfg.GNodeB.PlmnList.Mcc = "891"
	// cfg.GNodeB.PlmnList.Mnc = "23"
	// cfg.GNodeB.PlmnList.Tac = "000002"
	// cfg.GNodeB.SliceSupportList.St = "10"
	// cfg.GNodeB.SliceSupportList.Sst = "010239"

	// Ctrl-C interrupts a gNB that is still starting.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wg.Add(1)

	if _, err := gnb.InitGnb(ctx, cfg, &wg); err != nil {
		return err
	}

	wg.Wait()
	return nil
}
