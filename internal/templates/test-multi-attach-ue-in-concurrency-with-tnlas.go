package templates

import (
	"fmt"
	"my5G-RANTester/config"
	control_test_engine "my5G-RANTester/internal/control_test_engine"
	"my5G-RANTester/internal/data_test_engine"
	"sync"
	"time"
)

// testing attach and ping for a UE with TNLA.
func attachUeWithTnla(imsi string, conf config.Config, ranUeId int64, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(conf.AMF.Ip, conf.GNodeB.ControlIF.Ip, conf.AMF.Port, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// authentication to a GNB.
	gnbContext, err := control_test_engine.RegistrationGNB(conn, "000102", "free5gc", conf)
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	ue, err := control_test_engine.RegistrationUE(conn, imsi, ranUeId, conf, gnbContext, "208", "93")
	if err != nil {
		fmt.Println("The test failed when UE %s tried to attach! Error:%s", ue.Suci, err)
	}

	// end sctp socket.
	conn.Close()

	if err == nil {
		fmt.Println("Thread with imsi:%s and IP:%s worked fine", imsi, ue.UeIp)
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
	ranPort := cfg.GNodeB.ControlIF.Port

	// make n3(RAN connect to UPF)
	upfConn, err := data_test_engine.ConnectToUpf(cfg.GNodeB.DataIF.Ip, cfg.UPF.Ip, cfg.GNodeB.DataIF.Port, cfg.UPF.Port)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberUesConcurrency; i++ {
		imsi := control_test_engine.ImsiGenerator(i)
		wg.Add(1)
		go attachUeWithTnla(imsi, cfg, int64(i), &wg, ranPort)
		ranPort++
		// time.Sleep(10 * time.Millisecond)
	}

	// wait for multiple goroutines.
	wg.Wait()

	// check data plane UE.
	for i := 1; i <= numberUesConcurrency; i++ {
		// check data plane UE.
		gtpHeader := data_test_engine.GenerateGtpHeader(i)

		// ping test.
		err = data_test_engine.PingUE(upfConn, gtpHeader, data_test_engine.GetSrcPing(i), cfg.Ue.Ping)
		if err != nil {
			fmt.Println("The test failed when UE tried to use ping! Error:%s", err)
		}
		time.Sleep(1 * time.Second)
	}

	// function worked fine.
	return nil
}
