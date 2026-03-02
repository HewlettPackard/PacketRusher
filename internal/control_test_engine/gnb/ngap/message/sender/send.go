/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sender

import (
	"fmt"
	"my5G-RANTester/lib/ngap/ngapSctp"

	"github.com/ishidawataru/sctp"
)

func SendToAmF(message []byte, conn *sctp.SCTPConn) error {

	if conn == nil {
		return fmt.Errorf("SCTP connection is nil, cannot send NGAP message")
	}

	// TODO included information for SCTP association.
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}

	_, err := conn.SCTPWrite(message, info)
	if err != nil {
		return fmt.Errorf("error sending NGAP message %w", err)
	}

	return nil
}
