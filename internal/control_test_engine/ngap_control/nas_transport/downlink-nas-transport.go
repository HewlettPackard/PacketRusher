package nas_transport

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapType"
	"time"
)

func DownlinkNasTransport(connN2 *sctp.SCTPConn, supi string) (*ngapType.NGAPPDU, error) {

	// make channels
	c1 := make(chan error)
	c2 := make(chan *ngapType.NGAPPDU)

	// receive NGAP message from AMF.
	go func() {
		var recvMsg = make([]byte, 2048)
		var n int

		n, err := connN2.Read(recvMsg)
		if err != nil {
			c1 <- fmt.Errorf("Error receiving %s ue NGAP message in downlinkNasTransport", supi)
		}

		ngapMsg, err := ngap.Decoder(recvMsg[:n])
		if err != nil {
			c1 <- fmt.Errorf("Error decoding %s ue NGAP message in downlinkNasTransport", supi)
		}

		// worked fine.
		c2 <- ngapMsg
	}()

	// monitoring thread
	select {

	case msg1 := <-c1:
		// fmt.Println("Problem")
		return nil, msg1
	case msg2 := <-c2:
		// fmt.Println("OK")
		return msg2, nil
	case <-time.After(2 * time.Second):
		// fmt.Println("time")
		return nil, nil
	}
	// return ngapMsg, nil
}
