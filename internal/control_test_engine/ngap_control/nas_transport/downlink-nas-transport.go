package nas_transport

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
)

func DownlinkNasTransport(connN2 *sctp.SCTPConn) (*ngapType.NGAPPDU, error) {
	var recvMsg = make([]byte, 2048)
	var n int

	// receive NGAP message from AMF.
	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %s ue NGAP message in downlinkNasTransport")
	}

	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return nil, fmt.Errorf("Error decoding %s ue NGAP message in downlinkNasTransport")
	}

	return ngapMsg, fmt.Errorf("DowlinkNasTransport worked fine")
}
