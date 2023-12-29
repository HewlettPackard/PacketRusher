/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

// Package service
package service

import (
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

func InitConn(ue *context.UEContext, gnb *gnbContext.GNBContext) {
	inboundChannel := gnb.GetInboundChannel()

	ue.SetGnbRx(make(chan gnbContext.UEMessage, 1))
	ue.SetGnbTx(make(chan gnbContext.UEMessage, 1))

	// Send channels to gNB
	inboundChannel <- gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx(), PrUeId: ue.GetPrUeId()}
	msg := <-ue.GetGnbTx()
	ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
}
