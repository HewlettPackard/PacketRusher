package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/monitoring"
	"time"
)

func TestAvailability(interval int) {

	monitor := monitoring.Monitor{}

	conf, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

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
