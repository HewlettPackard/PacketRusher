/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package sender

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
)

func SendToUe(ue *context.GNBUe, message []byte) {
	SendMessageToUe(ue, context.UEMessage{IsNas: true, Nas: message})
}

func SendMessageToUe(ue *context.GNBUe, message context.UEMessage) {
	if !ue.DeliverToUE(message) {
		log.Warn("[GNB] Cannot send message to UE ", ue.GetRanUeId(), " as channel is closed")
	}
}
