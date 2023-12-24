/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"

	log "github.com/sirupsen/logrus"
)

func InitServer(gnb *context.GNBContext) {
	go gnbListen(gnb)
}

func gnbListen(gnb *context.GNBContext) {
	ln := gnb.GetInboundChannel()

	for {
		message := <-ln

		// TODO this region of the code may induces race condition.

		// new instance GNB UE context
		// store UE in UE Pool
		// store UE connection
		// select AMF and get sctp association
		// make a tun interface
		ue := gnb.NewGnBUe(message.GNBTx, message.GNBRx, message.PrUeId)
		if message.UEContext == nil {
			log.Info("[GNB] Received incoming connection from new UE")
			mcc, mnc := gnb.GetMccAndMnc()
			message.GNBTx <- context.UEMessage{Mcc: mcc, Mnc: mnc}
			ue.SetPduSessions(message.GNBPduSessions)
		} else {
			log.Info("[GNB] Received incoming handover for UE from another gNodeB")
			ue.SetStateReady()
			ue.CopyFromPreviousContext(message.UEContext)
			trigger.SendPathSwitchRequest(gnb, ue)
		}

		if ue == nil {
			log.Warn("[GNB] UE has not been created")
			break
		}

		// accept and handle connection.
		go processingConn(ue, gnb)
	}
}

func processingConn(ue *context.GNBUe, gnb *context.GNBContext) {
	rx := ue.GetGnbRx()
	for {
		message, done := <-rx
		gnbUeContext, err := gnb.GetGnbUe(ue.GetRanUeId())
		if (gnbUeContext == nil || err != nil) && done {
			log.Error("[GNB][NAS] Ignoring message from UE ", ue.GetRanUeId(), " as UE Context was cleaned as requested by AMF.")
			break
		}
		if !done {
			if gnbUeContext != nil {
				gnbUeContext.SetStateDown()
			}
			break
		}

		// send to dispatch.
		if message.ConnectionClosed {
			log.Info("[GNB] Cleaning up context on current gNb")
			gnbUeContext.SetStateDown()
			gnb.DeleteGnBUe(ue)
		} else if message.IsNas {
			nas.Dispatch(ue, message.Nas, gnb)
		} else {
			log.Error("[GNB] Received unknown message from UE")
		}
	}
}
