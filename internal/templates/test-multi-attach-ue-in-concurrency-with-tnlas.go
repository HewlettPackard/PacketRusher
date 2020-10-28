package templates

import (
	"fmt"
	"my5G-RANTester/config"
	control_test_engine "my5G-RANTester/internal/control_test_engine"
	"sync"
)

// testing attach and ping for a UE with TNLA.
func attachUeWithTnla(imsi string, ranUeId int64, amfIp string, ranIp string, amfPort int, ranIpAddr string, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(amfIp, ranIp, amfPort, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// authentication to a GNB.
	err = control_test_engine.RegistrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	suci, err := control_test_engine.RegistrationUE(conn, imsi, ranUeId, ranIpAddr)
	if err != nil {
		fmt.Println("The test failed when UE %s tried to attach! Error:%s", suci, err)
	}

	// end sockets.
	conn.Close()

	if err == nil {
		fmt.Println("Thread with imsi:%s worked fine", imsi)
	}
}

// testing attach and ping for multiple concurrent UEs using TNLAs.
func TestMultiAttachUesInConcurrencyWithTNLAs(numberUesConcurrency int) error {

	var wg sync.WaitGroup

	cfg, err := config.GetConfig()
	if err != nil {
		return nil
	}

	// authentication and ping to some concurrent UEs.
	fmt.Println("mytest: ", cfg.GNodeB.ControlIF.Ip, cfg.GNodeB.ControlIF.Port)

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberUesConcurrency; i++ {
		imsi := control_test_engine.ImsiGenerator(i)
		ranPort := cfg.GNodeB.ControlIF.Port
		wg.Add(1)
		go attachUeWithTnla(imsi, int64(i), cfg.AMF.Ip, cfg.GNodeB.ControlIF.Ip, cfg.AMF.Port, cfg.GNodeB.DataIF.Ip, &wg, ranPort)
		ranPort++
	}

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	return nil
}
