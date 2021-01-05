package main

import (
	"log"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
	"sync"
	"time"
)

func main() {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	go gnb.InitGnb(cfg)

	wg.Add(1)

	time.Sleep(20 * time.Millisecond)

	go ue.RegistrationUe(cfg, 1)

	wg.Add(1)

	wg.Wait()

}
