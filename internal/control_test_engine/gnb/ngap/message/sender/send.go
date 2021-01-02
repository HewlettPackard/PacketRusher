package sender

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToAmF(ue *context.GNBUe, message []byte) {

	conn := ue.GetSCTP()
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("Error sending NGAP message")
	}
}
