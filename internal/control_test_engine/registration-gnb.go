package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
	"my5G-RANTester/internal/control_test_engine/ngap_control/nas_transport"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId []byte, nameGNB string) error {

	// send NGSetup request msg.
	err := interface_management.NgSetupRequest(connN2, gnbId, 24, nameGNB)
	if err != nil {
		fmt.Println(err)
	}

	// receive NGSetupResponse Msg
	_, err = nas_transport.DownlinkNasTransport(connN2)
	if err != nil {
		fmt.Println(err)
	}

	// function worked fine.
	return nil
}
