package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
	"strconv"
	"sync"
	"time"
)

func TestMultiUesInQueue(numUes uint8) {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	go gnb.InitGnb(cfg)

	wg.Add(1)

	time.Sleep(60 * time.Millisecond)

	cfg.Ue.Imsi = imsiGenerator(1)
	go ue.RegistrationUe(cfg, 1)
	wg.Add(1)

	time.Sleep(30 * time.Second)

	cfg.Ue.Imsi = imsiGenerator(2)
	go ue.RegistrationUe(cfg, 2)
	wg.Add(1)

	wg.Wait()
}

func imsiGenerator(i int) string {

	var base string
	switch true {
	case i < 10:
		base = "imsi-208930000000"
	case i < 100:
		base = "imsi-20893000000"
	case i >= 100:
		base = "imsi-2089300000"
	}

	imsi := base + strconv.Itoa(i)
	return imsi
}
