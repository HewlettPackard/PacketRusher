package templates

import (
	"encoding/hex"
	"fmt"
	control_test_engine "my5G-RANTester/internal/control-test-engine"
)

// testing multiple GNBs authentication( control plane only)-> NGAP Request and response tester.
func testMultiAttachGnbInQueue(numberGnbs int) error {

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
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
		err = control_test_engine.RegistrationGNB(conn, resu, nameGNB)
		if err != nil {
			return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
		}
	}

	// functions worked fine.
	return nil
}
