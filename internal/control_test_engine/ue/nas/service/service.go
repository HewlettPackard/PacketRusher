/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package service
package service

import (
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"time"

	log "my5G-RANTester/internal/log"
)

func InitConn(ue *context.UEContext, gnbInboundChannel chan gnbContext.UEMessage) {
	ue.SetGnbRx(make(chan gnbContext.UEMessage, 10))
	ue.SetGnbTx(make(chan gnbContext.UEMessage, 10))
	lost := make(chan struct{})
	ue.SetGnbConnectionLost(lost)

	// Send channels to gNB
	gnbInboundChannel <- gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx(), PrUeId: ue.GetPrUeId(), Tmsi: ue.Get5gGuti(), ConnectionLost: lost}

	// Use timeout to prevent blocking indefinitely
	select {
	case msg, open := <-ue.GetGnbTx():
		if open {
			ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
		}
	case <-lost:
		log.Warn("[UE] gNB connection failed during attach")
	case <-time.After(5 * time.Second):
		log.Error("[UE] Timeout waiting for AMF MCC/MNC message from gNB")
	}
}
