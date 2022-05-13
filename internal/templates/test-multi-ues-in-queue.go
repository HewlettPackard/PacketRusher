package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
	"strconv"
	"sync"
	"time"
	"fmt"
)

func TestMultiUesInQueue(numUes int) {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	go gnb.InitGnb(cfg, &wg)

	wg.Add(1)

	time.Sleep(1 * time.Second)
    msin :=  cfg.Ue.Msin
	for i := 1; i <= numUes; i++ {

		imsi := imsiGenerator(i, msin)
		log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", imsi, " UE")
		cfg.Ue.Msin = imsi
		go ue.RegistrationUe(cfg, uint8(i), &wg)
		wg.Add(1)

		time.Sleep(10 * time.Second)
	}

	wg.Wait()
}

func imsiGenerator(i int, msin string) string {

	msin_int, err := strconv.Atoi(msin)
	if err != nil {
		log.Fatal("Error in get configuration")
	}
	base := msin_int + (i -1)

	imsi := fmt.Sprintf("%010d", base)
	return imsi
}