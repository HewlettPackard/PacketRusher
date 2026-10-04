/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/monitoring"
	"time"

	log "my5G-RANTester/internal/log"
)

func TestAvailability(interval int) {

	monitor := monitoring.Monitor{}

	conf := config.GetConfig()

	ranPort := uint16(1000)
	for y := 1; y <= interval; y++ {

		monitor.InitAvaibility()

		for i := 1; i <= 1; i++ {

			conf.GNodeB.ControlIF = conf.GNodeB.ControlIF.WithPort(ranPort)

			go gnb.InitGnbForAvaibility(conf, &monitor)

			ranPort++
		}

		time.Sleep(1020 * time.Millisecond)

		if monitor.GetAvailability() {
			log.Warn("[TESTER][GNB] AMF Availability:", 1)

		} else {
			log.Warn("[TESTER][GNB] AMF Availability:", 0)

		}
	}
}
