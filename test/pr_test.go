/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

	sender.Init(0)

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

	type UECheck struct {
		HasAuthOnce  bool
		PduActivated map[int32]bool
	}

	ueChecks := map[string]*UECheck{}

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithPDUCallback(context.Active, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			ue := args["ue"].(*context.UEContext)
			sm := args["sm"].(*context.SmContext)
			check := ueChecks[ue.GetSecurityContext().GetMsin()]
			if check.PduActivated == nil {
				check.PduActivated = map[int32]bool{}
			}
			check.PduActivated[sm.GetPduSessionId()] = true
		}).
		WithUeCallback(context.Authenticated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			ue := args["ue"].(*context.UEContext)
			check, ok := ueChecks[ue.GetSecurityContext().GetMsin()]
			if !ok {
				check = &UECheck{}
				ueChecks[ue.GetSecurityContext().GetMsin()] = check
			}
			check.HasAuthOnce = true
		}).
		Build()
	if err != nil {
		log.Printf("[5GC] Error during 5GC creation  %v", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)

	// Setup gNodeB
	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(gnbCount, conf, &wg)

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	// Setup UE
	ueCount := 10
	scenarioChans := make([]chan procedures.UeTesterMessage, ueCount+1)
	ueSimCfg := tools.UESimulationConfig{
		Gnbs:                     gnbs,
		Cfg:                      conf,
		TimeBeforeDeregistration: 400,
		TimeBeforeNgapHandover:   0,
		TimeBeforeXnHandover:     0,
		NumPduSessions:           1,
	}

	for ueSimCfg.UeId = 1; ueSimCfg.UeId <= ueCount; ueSimCfg.UeId++ {
		ueSimCfg.ScenarioChan = scenarioChans[ueSimCfg.UeId]

		imsi := tools.IncrementMsin(ueSimCfg.UeId, ueSimCfg.Cfg.Ue.Msin)

		securityContext := context.SecurityContext{}
		securityContext.SetMsin(imsi)
		securityContext.SetAuthSubscription(ueSimCfg.Cfg.Ue.Key, ueSimCfg.Cfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})

		amfContext := fiveGC.GetAMFContext()
		amfContext.Provision(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd}, securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	i := 0
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			i++
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			check := ueChecks[ue.GetSecurityContext().GetMsin()]
			assert.True(t, check.HasAuthOnce, "UE has never changed state")
			ue.ExecuteForAllSmContexts(
				func(sm *context.SmContext) {
					assert.Equalf(t, context.Inactive, sm.GetState().Current(), "Expected all pdu sessions to be in Inactive state but was not")
					assert.True(t, check.PduActivated[sm.GetPduSessionId()], "Expected all pdu to be activate once but was not")
				})
		})
	assert.Equalf(t, ueCount, i, "Expected %v ue to created in 5GC state but was %v", ueCount, i)

}

func TestUERegistrationLoop(t *testing.T) {
	controlIFConfig := config.ControlIF{
		Ip:   "127.0.0.1",
		Port: 9490,
	}
	dataIFConfig := config.DataIF{
		Ip:   "127.0.0.1",
		Port: 2155,
	}
	amfConfig := config.AMF{
		Ip:   "127.0.0.1",
		Port: 38415,
	}

	type UECheck struct {
		HasAuthOnce bool
	}
	ueChecks := map[string]*UECheck{}

	sender.Init(0)
	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfConfig)

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithUeCallback(context.Authenticated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			ue := args["ue"].(*context.UEContext)
			check, ok := ueChecks[ue.GetSecurityContext().GetMsin()]
			if !ok {
				check = &UECheck{}
				ueChecks[ue.GetSecurityContext().GetMsin()] = check
			}
			check.HasAuthOnce = true
		}).
		Build()
	if err != nil {
		log.Printf("[5GC] Error during 5GC creation  %v", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)

	// Setup gNodeB
	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(gnbCount, conf, &wg)

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	// Setup UE
	loopCount := 5
	scenarioChans := make([]chan procedures.UeTesterMessage, 2)
	ueSimCfg := tools.UESimulationConfig{
		UeId:                     1,
		Gnbs:                     gnbs,
		Cfg:                      conf,
		TimeBeforeDeregistration: 3000,
		TimeBeforeNgapHandover:   0,
		TimeBeforeXnHandover:     0,
		NumPduSessions:           1,
	}

	for i := 0; i < loopCount; i++ {
		if scenarioChans[ueSimCfg.UeId] != nil {
			scenarioChans[ueSimCfg.UeId] <- procedures.UeTesterMessage{Type: procedures.Kill}
			close(scenarioChans[ueSimCfg.UeId])
			scenarioChans[ueSimCfg.UeId] = nil
		}
		scenarioChans[ueSimCfg.UeId] = make(chan procedures.UeTesterMessage)
		ueSimCfg.ScenarioChan = scenarioChans[ueSimCfg.UeId]

		securityContext := context.SecurityContext{}
		securityContext.SetMsin(tools.IncrementMsin(ueSimCfg.UeId, ueSimCfg.Cfg.Ue.Msin))
		securityContext.SetAuthSubscription(ueSimCfg.Cfg.Ue.Key, ueSimCfg.Cfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})

		amfContext := fiveGC.GetAMFContext()
		amfContext.Provision(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd}, securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 2000 ms
		time.Sleep(time.Duration(2000) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			assert.True(t, ueChecks[ue.GetSecurityContext().GetMsin()].HasAuthOnce, "UE has never changed state")
		})
}
