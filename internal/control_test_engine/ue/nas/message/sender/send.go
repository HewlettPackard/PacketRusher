package sender

import (
	log "github.com/sirupsen/logrus"
	context2 "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func SendToGnb(ue *context.UEContext, message []byte) {
	ue.Lock()
	gnbRx := ue.GetGnbRx()
	if gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		gnbRx <- context2.UEMessage{IsNas: true, Nas: message}
	}
	ue.Unlock()
}