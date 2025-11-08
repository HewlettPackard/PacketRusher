/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sender

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	log "github.com/sirupsen/logrus"
)

func SendToUe(ue *context.GNBUe, message []byte) {
	ue.Lock()
	defer ue.Unlock()

	gnbTx := ue.GetGnbTx()
	if gnbTx == nil {
		log.Warn("[GNB] Cannot send NAS message to UE ", ue.GetRanUeId(), " as channel is closed")
		return
	}

	// Use non-blocking send to prevent deadlock during rapid operations
	select {
	case gnbTx <- context.UEMessage{IsNas: true, Nas: message}:
		log.Debug("[GNB] Successfully sent NAS message to UE ", ue.GetRanUeId())
	default:
		log.Warn("[GNB] Channel full, dropping NAS message for UE ", ue.GetRanUeId())
	}
}

func SendMessageToUe(ue *context.GNBUe, message context.UEMessage) {
	ue.Lock()
	defer ue.Unlock()

	gnbTx := ue.GetGnbTx()
	if gnbTx == nil {
		log.Warn("[GNB] Cannot send message to UE ", ue.GetRanUeId(), " as channel is closed")
		return
	}

	// Use non-blocking send to prevent deadlock during rapid operations
	select {
	case gnbTx <- message:
		log.Debug("[GNB] Successfully sent message to UE ", ue.GetRanUeId())
	default:
		log.Warn("[GNB] Channel full, dropping message for UE ", ue.GetRanUeId())
	}
}
