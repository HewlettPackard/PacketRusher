package nas_transport

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
	"time"
)

func DownlinkNasTransport(connN2 *sctp.SCTPConn, supi string) (*ngapType.NGAPPDU, error) {

	var recvMsg = make([]byte, 2048)
	var n int

	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %s ue NGAP message in downlinkNasTransport", supi)
	}

	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return nil, fmt.Errorf("Error decoding %s ue NGAP message in downlinkNasTransport", supi)
	}

	return ngapMsg, nil
}

func DownlinkNasTransportForConfigurationUpdateCommand(connN2 *sctp.SCTPConn, supi string) *ngapType.NGAPPDU {

	// make channels
	c1 := make(chan bool)
	c2 := make(chan *ngapType.NGAPPDU)

	// receive NGAP message from AMF.
	go func() {
		var recvMsg = make([]byte, 2048)
		var n int

		n, err := connN2.Read(recvMsg)
		if err != nil {
			c1 <- true
		}

		ngapMsg, err := ngap.Decoder(recvMsg[:n])
		if err != nil {
			c1 <- true
		}

		// worked fine.
		c2 <- ngapMsg
		log.WithFields(log.Fields{
			"protocol":    "ngap",
			"source":      "AMF",
			"destination": "gNodeB",
			"message":     "DownlinkNasTransport",
		}).Info("Receiving message")
	}()

	// monitoring thread
	select {

	case <-c1:
		fmt.Println("Error in receive configuration update command")
		break
	case <-c2:
		fmt.Println("Receive configuration update command")
		return <-c2
	case <-time.After(1000 * time.Millisecond):
		close(c1)
		close(c2)
		fmt.Println("timeout")
	}
	return nil
}
