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

func TestMultiUesInQueue(numUes int) {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	go gnb.InitGnb(cfg)

	wg.Add(1)

	time.Sleep(1 * time.Second)

	for i := 1; i <= numUes; i++ {

		imsi := imsiGenerator(i)
		log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", imsi, " UE")
		cfg.Ue.Msin = imsi
		go ue.RegistrationUe(cfg, uint8(i))
		wg.Add(1)

		time.Sleep(10 * time.Second)
	}

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
