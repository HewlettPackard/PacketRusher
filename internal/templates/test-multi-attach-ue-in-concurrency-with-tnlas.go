package templates

import (
	log "github.com/sirupsen/logrus"
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
	log.Info("Conecting to AMF...")
	conn, err := control_test_engine.ConnectToAmf(conf.AMF.Ip, conf.GNodeB.ControlIF.Ip, conf.AMF.Port, ranPort)
	if err != nil {
		log.Fatal("The test failed when sctp socket tried to connect to AMF! Error:", err)
	}
	log.Info("OK")

	// authentication to a GNB.
	gnbContext, err := control_test_engine.RegistrationGNB(conn, conf.GNodeB.PlmnList.GnbId, "my5GRANTester", conf)
	if err != nil {
		log.Fatal("The test failed when GNB tried to attach! Error:", err)
	}

	ue, err := control_test_engine.RegistrationUE(conn, imsi, ranUeId, conf, gnbContext, "208", "93")
	if err != nil {
		log.Error("The test failed when UE", ue.Suci, "tried to attach! Error:", err)
	}

	// end sctp socket.
	conn.Close()

}

// testing attach and ping for multiple concurrent UEs using TNLAs.
func TestMultiAttachUesInConcurrencyWithTNLAs(numberUesConcurrency int) {

	var wg sync.WaitGroup

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	// authentication and ping to some concurrent UEs.
	log.Info("Testing attached with ", numberUesConcurrency, " ues using TNLAs")
	log.Info("[CORE]", cfg.AMF.Name, " Core in Testing")

	ranPort := cfg.GNodeB.ControlIF.Port

	// make n3(RAN connect to UPF)
	log.Info("Conecting to UPF...")
	upfConn, err := data_test_engine.ConnectToUpf(cfg.GNodeB.DataIF.Ip, cfg.UPF.Ip, cfg.GNodeB.DataIF.Port, cfg.UPF.Port)
	if err != nil {
		log.Fatal("The test failed when udp socket tried to connect to UPF! Error:", err)
	}
	log.Info("OK")

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
			log.Error("The test failed when UE tried to use ping! Error:", err)
		}
		time.Sleep(1 * time.Second)
	}

}
