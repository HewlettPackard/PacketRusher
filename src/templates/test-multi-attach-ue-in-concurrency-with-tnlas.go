package templates

import (
	"fmt"
	control_test_engine "my5G-RANTester/src/control-test-engine"
	"sync"
)

// testing attach and ping for a UE with TNLA.
func attachUeWithTnla(ranUeId int64, ranIpAddr string, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// authentication to a GNB.
	err = control_test_engine.RegistrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a UE.
	// suciv1, suciv2, err := encodeUeSuci(imsi)
	//if err != nil {
	//fmt.Println("The test failed when SUCI was created! Error:%s", err)
	//}

	imsi, err := control_test_engine.RegistrationUE(conn, int(ranUeId), ranUeId, ranIpAddr)
	if err != nil {
		fmt.Println("The test failed when UE %s tried to attach! Error:%s", imsi, err)
	}

	// data plane UE
	// gtpHeader := generateGtpHeader(1)
	// err = pingUE(upfConn, gtpHeader, "60.60.0.1")
	// if err != nil {
	// return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
	//}

	// end sockets.
	conn.Close()
	//upfConn.Close()

	if err == nil {
		fmt.Println("Thread with imsi:%s worked fine", imsi)
	}
}

// testing attach and ping for multiple concurrent UEs using TNLAs.
func TestMultiAttachUesInConcurrencyWithTNLAs(numberUesConcurrency int) error {

	var wg sync.WaitGroup
	ranPort := 9487

	// authentication and ping to some concurrent UEs.

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberUesConcurrency; i++ {
		// imsi := generateImsi(i)
		wg.Add(1)
		go attachUeWithTnla(int64(i), "10.200.200.2", &wg, ranPort)
		ranPort++
	}

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	return nil
}
