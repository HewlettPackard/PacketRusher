package sender

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func SendToGnb(ue *context.UEContext, message []byte) {

	conn := ue.GetUnixConn()
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("Tratar o erro")
	}
}
