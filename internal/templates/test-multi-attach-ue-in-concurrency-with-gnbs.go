package templates

import (
	"fmt"
	control_test_engine "my5G-RANTester/internal/control-test-engine"
	"sync"
)

// testing attach and ping for multiple concurrent UEs using 2 GNBs.
func TestMultiAttachUesInConcurrencyWithGNBs() error {
	const ranIpAddr string = "10.200.200.2"
	const ran2IpAddr string = "10.200.200.1"

	var wg sync.WaitGroup

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	//upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// make N2(RAN2 connect to AMF)
	conn2, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, 9488)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket 2 tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN2 connect to UPF)
	//upfConn2, err := connectToUpf(ranIpAddr2, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// authentication to a GNB1.
	err = control_test_engine.RegistrationGNB(conn, []byte("\x00\x01\x01"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a GNB2.
	err = control_test_engine.RegistrationGNB(conn2, []byte("\x00\x01\x02"), "free5gc2")
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
			// imsi := generateImsi(i)

			// authentication to a UE.
			//suciv1, suciv2, err := encodeUeSuci(imsi)
			//if err != nil {
			//	fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			//}

			imsi, err := control_test_engine.RegistrationUE(conn, i, int64(i), ranIpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}
			// thread worked fine.
			// fmt.Println("Thread with imsi:%s worked fine", imsi)
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
			// imsi := generateImsi(i)

			// authentication to a UE.
			// suciv1, suciv2, err := encodeUeSuci(imsi)
			// if err != nil {
			//	fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			// }

			imsi, err := control_test_engine.RegistrationUE(conn2, i, int64(i), ranIpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}

			// thread worked fine.
			// fmt.Println("Thread with imsi:%s worked fine", imsi)
		}

		conn2.Close()

	}(&wg)

	// wait for multiple goroutines.
	wg.Wait()

	// upfConn.Close()
	// upfConn2.Close()

	return nil
}
