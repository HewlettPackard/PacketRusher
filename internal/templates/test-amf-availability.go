/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/monitoring"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestAvailability(interval int) {

	monitor := monitoring.Monitor{}

	conf := config.GetConfig()

	ranPort := 1000
	for y := 1; y <= interval; y++ {

		monitor.InitAvaibility()

		for i := 1; i <= 1; i++ {

			conf.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)

			conf.GNodeB.ControlIF.Port = ranPort

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

	return
}
