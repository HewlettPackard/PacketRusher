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

	// Send channels to gNB
	inboundChannel <- gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx(), Msin: ue.GetMsin()}
	msg := <-ue.GetGnbTx()
	ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
}
