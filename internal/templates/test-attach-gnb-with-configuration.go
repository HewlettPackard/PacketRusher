/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// wrong messages:
	// cfg.GNodeB.PlmnList.Mcc = "891"
	// cfg.GNodeB.PlmnList.Mnc = "23"
	// cfg.GNodeB.PlmnList.Tac = "000002"
	// cfg.GNodeB.SliceSupportList.St = "10"
	// cfg.GNodeB.SliceSupportList.Sst = "010239"

	wg.Add(1)
	node, err := gnb.InitGnbContext(ctx, cfg, &wg)
	if err != nil {
		return err
	}
	defer node.Terminate()

	wg.Wait()
	return nil
}
