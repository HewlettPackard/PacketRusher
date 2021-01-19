package sender

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToUe(ue *context.GNBUe, message []byte) {

	conn := ue.GetUnixSocket()
	_, err := conn.Write(message)
	if err != nil {
		log.Info("[GNB][UE] Error sending NAS message to UE")
	}
}
