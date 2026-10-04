/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package templates

import (
	"sync"

	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/monitoring"

	log "my5G-RANTester/internal/log"
)

// rajada de mensagens por segundo enviadas
// para o AMF
// durante um período de tempo
func TestRqsLoop(numRqs int, interval int) int64 {

	wg := sync.WaitGroup{}

	monitor := monitoring.Monitor{
		RqsL: 0,
		RqsG: 0,
	}

	cfg := config.GetConfig()
	identity := cfg.GNodeB.PlmnList
	if _, err := identity.GNBIDAt(max(numRqs, 1) - 1); err != nil {
		log.Fatalf("[TESTER][GNB] Invalid gNB identity range: %v", err)
	}

	ranPort := uint16(1000)
	for y := 1; y <= interval; y++ {

		monitor.InitRqsLocal()

		for i := 1; i <= numRqs; i++ {

			cfg.GNodeB.PlmnList.GnbId, _ = identity.GNBIDAt(i - 1)

			cfg.GNodeB.ControlIF = cfg.GNodeB.ControlIF.WithPort(ranPort)

			wg.Add(1)
			go gnb.InitGnbForLoadSeconds(cfg, &wg, &monitor)

			ranPort++
		}

		wg.Wait()

		log.Warn("[TESTER][GNB] AMF Responses per Second:", monitor.GetRqsLocal())
		monitor.SetRqsGlobal(monitor.GetRqsLocal())
	}

	return monitor.GetRqsGlobal()
}
