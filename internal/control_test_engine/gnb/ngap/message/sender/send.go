package sender

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToAmF(ue *context.GNBUe, message []byte) error {

	// TODO included information for SCTP association.

	conn := ue.GetSCTP()
	_, err := conn.Write(message)
	if err != nil {
		return fmt.Errorf("Error sending NGAP message")
	}
	return nil
}
