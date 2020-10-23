package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
	"my5G-RANTester/lib/ngap"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId []byte, nameGNB string) error {
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)
	var n int

	// send NGSetup request msg.
	sendMsg, err := interface_management.GetNGSetupRequest(gnbId, 24, nameGNB)
	if err != nil {
		fmt.Println("get NGSetupRequest Msg")
		return fmt.Errorf("Error getting GNB %s NGSetup Request Msg", nameGNB)
	}

	_, err = connN2.Write(sendMsg)
	if err != nil {
		fmt.Println("send NGSetupRequest Msg")
		return fmt.Errorf("Error sending GNB %s NGSetup Request Msg", nameGNB)
	}

	// receive NGSetupResponse Msg
	n, err = connN2.Read(recvMsg)
	if err != nil {
		fmt.Println("read NGSetupResponse Msg")
		return fmt.Errorf("Error reading GNB %s NGSetup Response Msg", nameGNB)
	}

	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		fmt.Println("decoder NGSetupResponse Msg")
		return fmt.Errorf("Error decoding GNB %s NGSetup Response Msg", nameGNB)
	}

	// function worked fine.
	return nil
}
