/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	engineUeContest "my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestCreatePDUSession(t *testing.T) {

	controlIFConfig := config.ControlIF{
		Ip:   "127.0.0.1",
		Port: 9489,
	}
	dataIFConfig := config.DataIF{
		Ip:   "127.0.0.1",
		Port: 2154,
	}
	amfConfig := config.AMF{
		Ip:   "127.0.0.1",
		Port: 38414,
	}

	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfConfig)

	// setup 5GC

	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.WithConfig(conf).Build()
	if err != nil {
		log.Printf("[5GC] Error during 5GC creation  %v", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)

	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(gnbCount, conf, &wg)

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	ueCfg := conf

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(ueCfg.Ue.Msin)
	securityContext.SetAuthSubscription(ueCfg.Ue.Key, ueCfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})
	amfContext := fiveGC.GetAMFContext()
	amfContext.NewSecurityContext(securityContext)

	ueId := 1
	ueRx := make(chan procedures.UeTesterMessage)
	ueTx := ue.NewUE(ueCfg, uint8(ueId), ueRx, gnbs[keys[0]], &wg)

	// setup some PacketRusher UE
	ueRx <- procedures.UeTesterMessage{Type: procedures.Registration}

	loop := true
	state := engineUeContest.MM5G_NULL
	for loop {
		msg := <-ueTx
		log.Info("[UE] Switched from state ", state, " to state ", msg.StateChange)
		switch msg.StateChange {
		case engineUeContest.MM5G_REGISTERED:
			if state != msg.StateChange {
				for i := 0; i < 1; i++ {
					ueRx <- procedures.UeTesterMessage{Type: procedures.NewPDUSession}
				}
			}
		case engineUeContest.MM5G_NULL:
			loop = false
		}
		state = msg.StateChange
	}

	wg.Wait()
	// assert.True(t, true)
}
