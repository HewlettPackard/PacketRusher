/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package interface_management

import (
	"fmt"

	"github.com/free5gc/ngap"
	"github.com/free5gc/ngap/ngapType"

	"github.com/ishidawataru/sctp"
)

func NgSetupResponse(connN2 *sctp.SCTPConn) (*ngapType.NGAPPDU, error) {
	var recvMsg = make([]byte, 2048)
	var n int

	// receive NGAP message from AMF.
	n, err := connN2.Read(recvMsg)
	if err != nil {
		return nil, fmt.Errorf("Error receiving %w NG-SETUP-RESPONSE", err)
	}

	ngapMsg, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		return nil, fmt.Errorf("Error decoding %w NG-SETUP-RESPONSE", err)
	}

	return ngapMsg, nil
}
