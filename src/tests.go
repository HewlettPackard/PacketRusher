package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// testing attach and ping for a UE with TNLA.
func AttachUeWithTnla(imsi string, ranUeId int64, ranIpAddr string, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a UE.
	suciv1, suciv2, err := encodeUeSuci(imsi)
	if err != nil {
		fmt.Println("The test failed when SUCI was created! Error:%s", err)
	}
	err = registrationUE(conn, imsi, ranUeId, suciv2, suciv1, ranIpAddr)
	if err != nil {
		fmt.Println("The test failed when UE tried to attach! Error:%s", err)
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

	fmt.Println("Thread with imsi:%s worked fine", imsi)
}

// testing authentication for a GNB( Control plane and User plane).
func testAttachGnb() error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	// function worked fine.
	return nil
}

// testing multiple GNBs authentication( control plane only)-> NGAP Request and response tester.
func testMultiAttachGnbInQueue(numberGnbs int) error {

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	for i := 1; i <= numberGnbs; i++ {

		// multiple names for GNBs.
		nameGNB := "my5gRanTester" + string(i)

		// generate GNB id.
		var aux string
		if i < 16 {
			aux = "00000" + fmt.Sprintf("%x", i)
		} else if i < 256 {
			aux = "0000" + fmt.Sprintf("%x", i)
		} else {
			aux = "000" + fmt.Sprintf("%x", i)
		}

		resu, err := hex.DecodeString(aux)
		if err != nil {
			return fmt.Errorf("error in GNB id for testing multiple GNBs")
		}

		// authentication to a GNB.
		err = registrationGNB(conn, resu, nameGNB)
		if err != nil {
			return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
		}
	}

	// functions worked fine.
	return nil
}

// testing multiple GNBs authentication(control plane only)-> NGAP Request and response tester.
func testMultiAttachGnbInConcurrency(numberGnbs int) error {

	ranPort := 9487
	var wg sync.WaitGroup

	// multiple concurrent GNBs authentication using goroutines.
	for i := 1; i <= numberGnbs; i++ {

		wg.Add(1)
		go func(wg *sync.WaitGroup, ranPort int, i int) {

			defer wg.Done()

			// make N2(RAN connect to AMF)
			conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, ranPort)
			if err != nil {
				fmt.Printf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
			}

			// multiple names for GNBs.
			nameGNB := "my5gRanTester" + string(i)

			// generate GNB id.
			var aux string
			if i < 16 {
				aux = "00000" + fmt.Sprintf("%x", i)
			} else if i < 256 {
				aux = "0000" + fmt.Sprintf("%x", i)
			} else {
				aux = "000" + fmt.Sprintf("%x", i)
			}

			resu, err := hex.DecodeString(aux)
			if err != nil {
				fmt.Printf("error in GNB id for testing multiple GNBs")
			}

			// authentication to a GNB.
			err = registrationGNB(conn, resu, nameGNB)
			if err != nil {
				fmt.Printf("The test failed when GNB tried to attach! Error:%s", err)
			}
		}(&wg, ranPort, i)
		ranPort++
	}

	// wait threads.
	wg.Wait()

	return nil
}

// testing attach and ping for multiple queued UEs.
func testMultiAttachUesInQueue(numberUes int) error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication and ping to some UEs.
	for i := 1; i <= numberUes; i++ {

		// generating some IMSIs to each UE.
		imsi := generateImsi(i)

		// authentication to a UE.
		suciv1, suciv2, err := encodeUeSuci(imsi)
		if err != nil {
			return fmt.Errorf("The test failed when SUCI was created! Error:%s", err)
		}

		err = registrationUE(conn, imsi, int64(i), suciv2, suciv1, ranIpAddr)
		if err != nil {
			return fmt.Errorf("The test failed when UE tried to attach! Error:%s", err)
		}

		// data plane UE
		ipUe := getSrcPing(i)
		gtpHeader := generateGtpHeader(i)

		err = pingUE(upfConn, gtpHeader, ipUe)
		if err != nil {
			return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
		}
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	return nil
}

// testing attach and ping for multiple concurrent UEs using 2 GNBs.
func testMultiAttachUesInConcurrencyWithGNBs() error {
	const ranIpAddr string = "10.200.200.2"
	const ran2IpAddr string = "10.200.200.1"

	var wg sync.WaitGroup

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	//upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// make N2(RAN2 connect to AMF)
	conn2, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9488)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket 2 tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN2 connect to UPF)
	//upfConn2, err := connectToUpf(ranIpAddr2, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// authentication to a GNB1.
	err = registrationGNB(conn, []byte("\x00\x01\x01"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a GNB2.
	err = registrationGNB(conn2, []byte("\x00\x01\x02"), "free5gc2")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication and ping to some concurrent UEs.

	// Launch several goroutines and increment the WaitGroup counter for each.
	wg.Add(1)

	// goroutine.
	go func(wg *sync.WaitGroup) {

		defer wg.Done()

		for i := 1; i <= 3; i++ {

			// generating some IMSIs to each UE.
			imsi := generateImsi(i)

			// authentication to a UE.
			suciv1, suciv2, err := encodeUeSuci(imsi)
			if err != nil {
				fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			}

			err = registrationUE(conn, imsi, int64(i), suciv2, suciv1, ranIpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}
			// thread worked fine.
			fmt.Println("Thread with imsi:%s worked fine", imsi)
		}

		conn.Close()

	}(&wg)

	// increment the WaitGroup counter.
	wg.Add(1)

	// goroutine.
	go func(wg *sync.WaitGroup) {

		defer wg.Done()

		for i := 4; i <= 6; i++ {
			// generating some IMSIs to each UE.
			imsi := generateImsi(i)

			// authentication to a UE.
			suciv1, suciv2, err := encodeUeSuci(imsi)
			if err != nil {
				fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			}

			err = registrationUE(conn2, imsi, int64(i), suciv2, suciv1, ran2IpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}

			// thread worked fine.
			fmt.Println("Thread with imsi:%s worked fine", imsi)
		}

		conn2.Close()

	}(&wg)

	// wait for multiple goroutines.
	wg.Wait()

	// upfConn.Close()
	// upfConn2.Close()

	return nil
}

// testing attach and ping for multiple concurrent UEs using TNLAs.
func testMultiAttachUesInConcurrencyWithTNLAs(numberUesConcurrency int) error {

	var wg sync.WaitGroup
	ranPort := 9487

	// authentication and ping to some concurrent UEs.

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= numberUesConcurrency; i++ {
		imsi := generateImsi(i)
		wg.Add(1)
		go AttachUeWithTnla(imsi, int64(i), "10.200.200.2", &wg, ranPort)
		ranPort++
	}

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	return nil
}
