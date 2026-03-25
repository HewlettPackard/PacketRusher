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
	gnbTx := ue.GetGnbTx()
	ue.Unlock()

	if gnbTx == nil {
		log.Warn("[GNB] Cannot send NAS message to UE ", ue.GetRanUeId(), " as channel is closed")
		return
	}

	// Block until there is space in the channel. Drop only if DeleteGnBUe closes
	// the channel concurrently (detected via panic recovery).
	defer func() {
		if r := recover(); r != nil {
			log.Warn("[GNB] NAS message dropped for UE ", ue.GetRanUeId(), ": channel closed concurrently")
		}
	}()
	gnbTx <- context.UEMessage{IsNas: true, Nas: message}
}

func SendMessageToUe(ue *context.GNBUe, message context.UEMessage) {
	ue.Lock()
	gnbTx := ue.GetGnbTx()
	ue.Unlock()

	if gnbTx == nil {
		log.Warn("[GNB] Cannot send message to UE ", ue.GetRanUeId(), " as channel is closed")
		return
	}

	// Block until there is space in the channel. Drop only if DeleteGnBUe closes
	// the channel concurrently (detected via panic recovery).
	defer func() {
		if r := recover(); r != nil {
			log.Warn("[GNB] Message dropped for UE ", ue.GetRanUeId(), ": channel closed concurrently")
		}
	}()
	gnbTx <- message
}
