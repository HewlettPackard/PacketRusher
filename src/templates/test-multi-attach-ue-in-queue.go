package templates

import (
	"fmt"
	control_test_engine "my5G-RANTester/src/control-test-engine"
	"my5G-RANTester/src/data-test-engine"
)

// testing attach and ping for multiple queued UEs.
func TestMultiAttachUesInQueue(numberUes int) error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := data_test_engine.ConnectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = control_test_engine.RegistrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication and ping to some UEs.
	for i := 1; i <= numberUes; i++ {

		// generating some IMSIs to each UE.
		// imsi := generateImsi(i)

		// authentication to a UE.
		// suciv1, suciv2, err := encodeUeSuci(imsi)
		// if err != nil {
		//	return fmt.Errorf("The test failed when SUCI was created! Error:%s", err)
		// }

		imsi, err := control_test_engine.RegistrationUE(conn, i, int64(i), ranIpAddr)
		if err != nil {
			return fmt.Errorf("The test failed when UE %s tried to attach! Error:%s", imsi, err)
		}

		// data plane UE
		// ipUe := getSrcPing(i)
		gtpHeader := data_test_engine.GenerateGtpHeader(i)

		err = data_test_engine.PingUE(upfConn, gtpHeader, i)
		if err != nil {
			return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
		}
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	return nil
}
