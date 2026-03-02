/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

// Package service
package service

import (
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"time"

	log "github.com/sirupsen/logrus"
)

func InitConn(ue *context.UEContext, gnbInboundChannel chan gnbContext.UEMessage) {
	ue.SetGnbRx(make(chan gnbContext.UEMessage, 10))
	ue.SetGnbTx(make(chan gnbContext.UEMessage, 10))

	// Send channels to gNB
	gnbInboundChannel <- gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx(), PrUeId: ue.GetPrUeId(), Tmsi: ue.Get5gGuti()}

	// Use timeout to prevent blocking indefinitely
	select {
	case msg := <-ue.GetGnbTx():
		ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
	case <-time.After(5 * time.Second):
		log.Error("[UE] Timeout waiting for AMF MCC/MNC message from gNB")
	}
}
