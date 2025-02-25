/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ue

import (
	"my5G-RANTester/config"
	context2 "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	serviceGtp "my5G-RANTester/internal/control_test_engine/ue/gtp/service"
	"my5G-RANTester/internal/control_test_engine/ue/nas/service"
	"my5G-RANTester/internal/control_test_engine/ue/nas/trigger"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"my5G-RANTester/internal/control_test_engine/ue/state"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func NewUE(conf config.Config, id int, ueMgrChannel chan procedures.UeTesterMessage, gnbInboundChannel chan context2.UEMessage, wg *sync.WaitGroup) chan scenario.ScenarioMessage {
	// new UE instance.
	ue := &context.UEContext{}
	scenarioChan := make(chan scenario.ScenarioMessage)

	// new UE context
	ue.NewRanUeContext(
		conf.Ue.Msin,
		conf.GetUESecurityCapability(),
		conf.Ue.Key,
		conf.Ue.Opc,
		"c9e8763286b5b9ffbdf56e1297d0887b",
		conf.Ue.Amf,
		conf.Ue.Sqn,
		conf.Ue.Hplmn.Mcc,
		conf.Ue.Hplmn.Mnc,
		conf.GetHomeNetworkPublicKey(),
		conf.Ue.RoutingIndicator,
		conf.Ue.Dnn,
		int32(conf.Ue.Snssai.Sst),
		conf.Ue.Snssai.Sd,
		conf.Ue.TunnelMode,
		scenarioChan,
		gnbInboundChannel,
		id)

	go func() {
		// starting communication with GNB and listen.
		service.InitConn(ue, ue.GetGnbInboundChannel())
		sigStop := make(chan os.Signal, 1)
		signal.Notify(sigStop, os.Interrupt)

		// Block until a signal is received.
		loop := true
		for loop {
			select {
			case msg, open := <-ue.GetGnbTx():
				if !open {
					log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with gNB was closed")
					ue.SetGnbTx(nil)
					break
				}
				gnbMsgHandler(msg, ue)
			case msg, open := <-ueMgrChannel:
				if !open {
					log.Warn("[UE][", ue.GetMsin(), "] Stopping UE as communication with scenario was closed")
					loop = false
					break
				}
				loop = ueMgrHandler(msg, ue)
			case <-ue.GetDRX():
				verifyPaging(ue)
			}
		}
		ue.Terminate()
		wg.Done()
	}()

	return scenarioChan
}

func gnbMsgHandler(msg context2.UEMessage, ue *context.UEContext) {
	if msg.IsNas {
		state.DispatchState(ue, msg.Nas)
	} else if msg.GNBPduSessions[0] != nil {
		// Setup PDU Session
		serviceGtp.SetupGtpInterface(ue, msg)
	} else if msg.GNBRx != nil && msg.GNBTx != nil && msg.GNBInboundChannel != nil {
		log.Info("[UE] gNodeB is telling us to use another gNodeB")
		previousGnbRx := ue.GetGnbRx()
		ue.SetGnbInboundChannel(msg.GNBInboundChannel)
		ue.SetGnbRx(msg.GNBRx)
		ue.SetGnbTx(msg.GNBTx)
		previousGnbRx <- context2.UEMessage{ConnectionClosed: true}
		close(previousGnbRx)
	} else {
		log.Error("[UE] Received unknown message from gNodeB", msg)
	}
}

func verifyPaging(ue *context.UEContext) {
	gnbTx := make(chan context2.UEMessage, 1)

	ue.GetGnbInboundChannel() <- context2.UEMessage{GNBTx: gnbTx, FetchPagedUEs: true}
	msg := <-gnbTx
	for _, pagedUE := range msg.PagedUEs {
		if ue.Get5gGuti() != nil && pagedUE.FiveGSTMSI != nil && [4]uint8(pagedUE.FiveGSTMSI.FiveGTMSI.Value) == ue.GetTMSI5G() {
			ueMgrHandler(procedures.UeTesterMessage{Type: procedures.ServiceRequest}, ue)
			return
		}
	}
}

func ueMgrHandler(msg procedures.UeTesterMessage, ue *context.UEContext) bool {
	loop := true
	switch msg.Type {
	case procedures.Registration:
		trigger.InitRegistration(ue)
	case procedures.Deregistration:
		trigger.InitDeregistration(ue)
	case procedures.NewPDUSession:
		trigger.InitPduSessionRequest(ue)
	case procedures.DestroyPDUSession:
		pdu, err := ue.GetPduSession(msg.Param)
		if err != nil {
			log.Error("[UE] Cannot release unknown PDU Session ID ", msg.Param)
			return loop
		}
		trigger.InitPduSessionRelease(ue, pdu)
	case procedures.Idle:
		// We switch UE to IDLE
		ue.SetStateMM_IDLE()
		trigger.SwitchToIdle(ue)
		ue.CreateDRX(25 * time.Millisecond)
	case procedures.ServiceRequest:
		if ue.GetStateMM() == context.MM5G_IDLE {
			ue.StopDRX()

			// Since gNodeB stopped communication after switching to Idle, we need to connect back to gNodeB
			service.InitConn(ue, ue.GetGnbInboundChannel())
			if ue.Get5gGuti() != nil {
				trigger.InitServiceRequest(ue)
			} else {
				// If AMF did not assign us a GUTI, we have to fallback to the usual Registration/Authentification process
				// PDU Sessions will still be recovered
				trigger.InitRegistration(ue)
			}
		}
	case procedures.Terminate:
		log.Info("[UE] Terminating UE as requested")
		// If UE is registered
		if ue.GetStateMM() == context.MM5G_REGISTERED {
			// Release PDU Sessions
			for i := uint8(1); i <= 16; i++ {
				pduSession, _ := ue.GetPduSession(i)
				if pduSession != nil {
					trigger.InitPduSessionRelease(ue, pduSession)
					select {
					case <-pduSession.Wait:
					case <-time.After(500 * time.Millisecond):
						// If still unregistered after 500 ms, continue
					}
				}
			}
			// Initiate Deregistration
			trigger.InitDeregistration(ue)
		}
		// Else, nothing to do
		loop = false
	case procedures.Kill:
		loop = false
	}
	return loop
}
