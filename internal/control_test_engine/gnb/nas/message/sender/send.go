package sender

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToUe(ue *context.GNBUe, message []byte) {
	conn := ue.GetGnbTx()
	conn <- context.UEMessage{IsNas: true, Nas: message}
}

func SendMessageToUe(ue *context.GNBUe, message context.UEMessage) {
	ue.GetGnbTx() <- message
}