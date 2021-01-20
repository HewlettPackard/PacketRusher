package interface_management

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
)

func NgSetupResponse(connN2 *sctp.SCTPConn) (*ngapType.NGAPPDU, error) {
	var recvMsg = make([]byte, 2048)
	var n int

	// receive NGAP message from AMF.
	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %s NG-SETUP-RESPONSE")
	}

	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return nil, fmt.Errorf("Error decoding %s NG-SETUP-RESPONSE")
	}

	return ngapMsg, nil
}
