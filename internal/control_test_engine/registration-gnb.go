package control_test_engine

import (
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
	"my5G-RANTester/internal/logging"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId []byte, nameGNB string) error {

	// send NGSetup request msg.
	err := interface_management.NgSetupRequest(connN2, gnbId, 24, nameGNB)
	if logging.Check_error(err, "send NGSetupRequest Message") {
		return err
	}

	// receive NGSetupResponse Msg
	_, err = interface_management.NgSetupResponse(connN2)
	if logging.Check_error(err, "receive NGSetupResponse Message") {
		return err
	}

	// function worked fine.
	return nil
}
