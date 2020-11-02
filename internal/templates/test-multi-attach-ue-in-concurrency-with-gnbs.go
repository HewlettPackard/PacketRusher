package templates

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine"
	"sync"
)

func attachUeWithGNB(imsi string, ranUeId int64, amfIp string, ranIp string, amfPort int, ranIpAddr string, wg *sync.WaitGroup, ranPort int, k string, opc string, amf string) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(amfIp, ranIp, amfPort, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
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

	resu, err := hex.DecodeString(aux)
	if err != nil {
		fmt.Printf("error in GNB id for testing multiple GNBs")
	}

	// authentication to a GNB.
	err = control_test_engine.RegistrationGNB(conn, resu, nameGNB)
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	suci, err := control_test_engine.RegistrationUE(conn, imsi, ranUeId, ranIpAddr, k, opc, amf)
	if err != nil {
		fmt.Println("The test failed when UE %s tried to attach! Error:%s", suci, err)
	}

	// end sockets.
	conn.Close()

	if err == nil {
		fmt.Println("Thread with imsi:%s worked fine", imsi)
	}
}

// testing attach for multiple concurrent UEs using an UE per GNB.
func TestMultiAttachUesInConcurrencyWithGNBs(numberGNBs int) error {

	var wg sync.WaitGroup

	cfg, err := config.GetConfig()
	if err != nil {
		return nil
	}

	fmt.Println("mytest: ", cfg.GNodeB.ControlIF.Ip, cfg.GNodeB.ControlIF.Port)

	ranPort := cfg.GNodeB.ControlIF.Port

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberGNBs; i++ {
		imsi := control_test_engine.ImsiGenerator(i)
		wg.Add(1)
		go attachUeWithGNB(imsi, int64(i), cfg.AMF.Ip, cfg.GNodeB.ControlIF.Ip, cfg.AMF.Port, cfg.GNodeB.DataIF.Ip, &wg, ranPort, cfg.Ue.Key, cfg.Ue.Opc, cfg.Ue.Amf)
		ranPort++
		// time.Sleep(10* time.Millisecond)
	}

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	return nil
}
