/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package nas_transport

import (
	"fmt"
	"time"

	ngap "github.com/free5gc/ngap/message"

	"github.com/ishidawataru/sctp"
	log "my5G-RANTester/internal/log"
)

func DownlinkNasTransport(connN2 *sctp.SCTPConn, supi string) (ngap.Message, error) {

	var recvMsg = make([]byte, 2048)
	var n int

	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %s ue NGAP message in downlinkNasTransport", supi)
	}

	ngapMsg, err := ngap.Parse(recvMsg[:n])
	if err != nil {
		return nil, fmt.Errorf("Error decoding %s ue NGAP message in downlinkNasTransport", supi)
	}

	return ngapMsg, nil
}

func DownlinkNasTransportForConfigurationUpdateCommand(connN2 *sctp.SCTPConn, supi string) ngap.Message {

	// make channels
	c1 := make(chan bool)
	c2 := make(chan ngap.Message)

	// receive NGAP message from AMF.
	go func() {
		var recvMsg = make([]byte, 2048)
		var n int

		n, err := connN2.Read(recvMsg)
		if err != nil {
			c1 <- true
		}

		ngapMsg, err := ngap.Parse(recvMsg[:n])
		if err != nil {
			c1 <- true
		}

		// worked fine.
		c2 <- ngapMsg
		log.Info("[GNB][NGAP] Receiving DownlinkNasTransport from AMF")
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
