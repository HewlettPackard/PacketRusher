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
	lost := make(chan struct{})
	ue.SetGnbConnectionLost(lost)

	// Only the owning gNB can order admission against closing its channel.
	gnb := ue.GetGnbContext()
	if gnb == nil || gnb.GetInboundChannel() != gnbInboundChannel {
		close(lost)
		log.Warn("[UE] Connection has no matching gNB lifecycle owner")
		return
	}
	if err := gnb.QueueUE(gnbContext.UEMessage{GNBTx: ue.GetGnbTx(), GNBRx: ue.GetGnbRx(), PrUeId: ue.GetPrUeId(), Tmsi: ue.Get5gGuti(), ConnectionLost: lost}); err != nil {
		// Rejected admission transfers no channels to the gNB; this UE owns loss.
		close(lost)
		log.Warn("[UE] gNB rejected connection: ", err)
		return
	}

	// Use timeout to prevent blocking indefinitely
	select {
	case msg, open := <-ue.GetGnbTx():
		if open {
			ue.SetAmfMccAndMnc(msg.Mcc, msg.Mnc)
		}
	case <-gnb.Done():
		log.Warn("[UE] gNB stopped during attach")
	case <-lost:
		log.Warn("[UE] gNB connection failed during attach")
	case <-time.After(5 * time.Second):
		log.Error("[UE] Timeout waiting for AMF MCC/MNC message from gNB")
	}
}
