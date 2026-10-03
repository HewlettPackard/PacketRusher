/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"sync"

	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/monitoring"

	log "github.com/sirupsen/logrus"
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
	if numRqs < 1 {
		return 0
	}
	if _, err := identity.GNBIDAt(numRqs - 1); err != nil {
		log.Errorf("[TESTER][GNB] Invalid gNB identity range: %v", err)
		return 0
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
