package sender

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToUe(ue *context.GNBUe, message []byte) {

	conn := ue.GetUnixSocket()
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("Erro sending NAS message to UE")
	}
}
