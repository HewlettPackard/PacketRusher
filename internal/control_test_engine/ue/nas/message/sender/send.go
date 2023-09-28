package sender

import (
	context2 "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func SendToGnb(ue *context.UEContext, message []byte) {

	conn := ue.GetGnbRx()
	conn <- context2.UEMessage{IsNas: true, Nas: message}
}
