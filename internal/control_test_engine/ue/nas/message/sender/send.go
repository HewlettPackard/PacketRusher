/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sender

import (
	context2 "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"

	log "github.com/sirupsen/logrus"
)

func SendToGnb(ue *context.UEContext, message []byte) {
	SendToGnbMsg(ue, context2.UEMessage{IsNas: true, Nas: message})
}

func SendToGnbMsg(ue *context.UEContext, message context2.UEMessage) {
	ue.Lock()
	gnbRx := ue.GetGnbRx()
	if gnbRx == nil {
		log.Warn("[UE] Do not send NAS messages to gNB as channel is closed")
	} else {
		gnbRx <- message
	}
	ue.Unlock()
}
