package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
)

func TestAttachGnbWithConfiguration() {

	// wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	// wrong messages:
	// cfg.GNodeB.PlmnList.Mcc = "891"
	// cfg.GNodeB.PlmnList.Mnc = "23"
	// cfg.GNodeB.PlmnList.Tac = "000002"
	// cfg.GNodeB.SliceSupportList.St = "10"
	// cfg.GNodeB.SliceSupportList.Sst = "010239"

	gnb.InitGnb(cfg)

	// wg.Add(1)

	// wg.Wait()
}
