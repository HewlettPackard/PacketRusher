// Package service
package service

import (
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func InitConn(ue *context.UEContext, gnb *gnbContext.GNBContext) chan gnbContext.UEMessage {
	inboundChannel := gnb.GetInboundChannel()

	// Send channels to gNB
	inboundChannel <- gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx()}

	return ue.GetGnbTx()
}
