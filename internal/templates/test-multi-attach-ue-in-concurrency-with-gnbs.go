package templates

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine"
	"sync"
)

func attachUeWithGNB(imsi string, conf config.Config, ranUeId int64, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(conf.AMF.Ip, conf.GNodeB.ControlIF.Ip, conf.AMF.Port, ranPort)
	if err != nil {
		log.Fatal("The test failed when sctp socket tried to connect to AMF! Error:", err)
	}

	// multiple names for GNBs.
	nameGNB := fmt.Sprint("my5gRanTester", ranUeId)

	// generate GNB id.
	var aux string
	if ranUeId < 16 {
		aux = "00000" + fmt.Sprintf("%x", ranUeId)
	} else if ranUeId < 256 {
		aux = "0000" + fmt.Sprintf("%x", ranUeId)
	} else {
		aux = "000" + fmt.Sprintf("%x", ranUeId)
	}

	// authentication to a GNB.
	gnbContext, err := control_test_engine.RegistrationGNB(conn, aux, nameGNB, conf)
	if err != nil {
		log.Fatal("The test failed when GNB tried to attach! Error:", err)
	}

	ue, err := control_test_engine.RegistrationUE(conn, imsi, ranUeId, conf, gnbContext, "208", "93")
	if err != nil {
		log.Error("The test failed when UE", ue.Suci, "tried to attach! Error:", err)
	}

	// end sockets.
	conn.Close()

	//if err == nil {
	//fmt.Println("Thread with imsi:%s and Ip: %s worked fine", imsi, ue.GetIp())
	//}
}

// testing attach for multiple concurrent UEs using an UE per GNB.
func TestMultiAttachUesInConcurrencyWithGNBs(numberGNBs int) {

	var wg sync.WaitGroup

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Error in get configuration")
	}

	log.Info("Testing attached with ", numberGNBs, " ues in different GNBs")
	log.Info("[CORE]", cfg.AMF.Name, " Core in Testing")

	ranPort := cfg.GNodeB.ControlIF.Port

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberGNBs; i++ {
		imsi := control_test_engine.ImsiGenerator(i)
		wg.Add(1)
		go attachUeWithGNB(imsi, cfg, int64(i), &wg, ranPort)
		ranPort++
		// time.Sleep(10* time.Millisecond)
	}

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	//return nil
}
