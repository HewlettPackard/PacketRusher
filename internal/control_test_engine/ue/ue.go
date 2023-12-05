/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ue

import (
	log "github.com/sirupsen/logrus"
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
)

func NewUE(conf config.Config, id uint8, ueMgrChannel chan procedures.UeTesterMessage, gnb *context2.GNBContext, wg *sync.WaitGroup) chan scenario.ScenarioMessage {
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
		conf.Ue.RoutingIndicator,
		conf.Ue.Dnn,
		int32(conf.Ue.Snssai.Sst),
		conf.Ue.Snssai.Sd,
		conf.Ue.TunnelEnabled,
		scenarioChan,
		id)

	go func() {
		// starting communication with GNB and listen.
		service.InitConn(ue, gnb)
		sigStop := make(chan os.Signal, 1)
		signal.Notify(sigStop, os.Interrupt)

		// Block until a signal is received.
		loop := true
		for loop {
			select {
			case msg, open := <-ue.GetGnbTx():
				if !open {
					log.Error("[UE][", ue.GetMsin(), "] Stopping UE as communication with gNB was closed")
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
			}
		}
		ue.Terminate()
		wg.Done()
	}()

	return scenarioChan
}

func gnbMsgHandler(msg context2.UEMessage, ue *context.UEContext) {
	if msg.IsNas {
		// handling NAS message.
		ue.SetAmfUeId(msg.AmfId)
		state.DispatchState(ue, msg.Nas)
	} else if msg.GNBPduSessions[0] != nil {
		// Setup PDU Session
		serviceGtp.SetupGtpInterface(ue, msg)
	} else {
		log.Error("[UE] Received unknown message from gNodeB", msg)
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
		if err == nil {
			log.Error("[UE] Cannot release unknown PDU Session ID ", msg.Param)
			return loop
		}
		trigger.InitPduSessionRelease(ue, pdu)
	case procedures.Handover:
		var pduSession *context.UEPDUSession
		for i := uint8(1); i <= 16; i++ {
			pduSession, _ = ue.GetPduSession(i)
			if pduSession != nil {
				break
			}
		}
		if pduSession == nil {
			log.Error("[UE] Cannot handover / PathSwitchRequest to a new gNodeB without any PDU Sessions")
			break
		}
		trigger.InitHandover(ue, msg.GnbChan)
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
					case <-time.After(5 * time.Millisecond):
						// If still unregistered after 5 ms, continue
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