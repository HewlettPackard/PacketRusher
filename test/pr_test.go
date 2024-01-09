/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/state"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/util/fsm"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

	f, err := fsm.NewFSM(
		fsm.Transitions{
			{Event: fsm.EventType(nas.MsgTypeRegistrationRequest), From: state.Fresh, To: state.RegistrationRequested},
			{Event: fsm.EventType(nas.MsgTypeAuthenticationResponse), From: state.RegistrationRequested, To: state.AuthenticationChallengeResponded},
			{Event: fsm.EventType(nas.MsgTypeSecurityModeComplete), From: state.AuthenticationChallengeResponded, To: state.SecurityContextSet},
			{Event: fsm.EventType(nas.MsgTypeRegistrationComplete), From: state.SecurityContextSet, To: state.Registred},
			{Event: fsm.EventType(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration), From: state.Registred, To: state.DeregistrationRequested},
			{Event: fsm.EventType(fmt.Sprint(ngapType.ProcedureCodeUEContextRelease)), From: state.DeregistrationRequested, To: state.Idle},
		},
		fsm.Callbacks{
			state.Fresh:                            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.RegistrationRequested:            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.AuthenticationChallengeResponded: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.SecurityContextSet:               func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.Registred:                        func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.DeregistrationRequested:          func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.Idle: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
				assert.Equal(t, args["requestedPduCount"], args["providedPduCount"], "Number of requested PDU Session count does not match the provided count")
				assert.Equal(t, args["providedPduCount"], args["releasedPduCount"], "Number of provided PDU Session count does not match the released count")
			}},
	)
	assert.Nil(t, err, "FSM creation failed")

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

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithFSM(f).
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

		securityContext := context.SecurityContext{}
		securityContext.SetMsin(tools.IncrementMsin(ueSimCfg.UeId, ueSimCfg.Cfg.Ue.Msin))
		securityContext.SetAuthSubscription(ueSimCfg.Cfg.Ue.Key, ueSimCfg.Cfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})
		securityContext.SetDefaultSNssai(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd})

		amfContext := fiveGC.GetAMFContext()
		amfContext.NewSecurityContext(securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			assert.Equalf(t, state.Idle, ue.GetState().GetState().Current(), "Expected all ue to be in Idle state but was not")
		})
}

func TestUERegistrationLoop(t *testing.T) {

	f, err := fsm.NewFSM(
		fsm.Transitions{
			{Event: fsm.EventType(nas.MsgTypeRegistrationRequest), From: state.Fresh, To: state.RegistrationRequested},
			{Event: fsm.EventType(nas.MsgTypeAuthenticationResponse), From: state.RegistrationRequested, To: state.AuthenticationChallengeResponded},
			{Event: fsm.EventType(nas.MsgTypeSecurityModeComplete), From: state.AuthenticationChallengeResponded, To: state.SecurityContextSet},
			{Event: fsm.EventType(nas.MsgTypeRegistrationComplete), From: state.SecurityContextSet, To: state.Registred},
			{Event: fsm.EventType(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration), From: state.Registred, To: state.DeregistrationRequested},
			{Event: fsm.EventType(fmt.Sprint(ngapType.ProcedureCodeUEContextRelease)), From: state.DeregistrationRequested, To: state.Idle},
			{Event: fsm.EventType(fmt.Sprint(ngapType.ProcedureCodeUEContextRelease)), From: state.Registred, To: state.Idle},
		},
		fsm.Callbacks{
			state.Fresh:                            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.RegistrationRequested:            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.AuthenticationChallengeResponded: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.SecurityContextSet:               func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.Registred:                        func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.DeregistrationRequested:          func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			state.Idle: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
				assert.Equal(t, args["requestedPduCount"], args["providedPduCount"], "Number of requested PDU Session count does not match the provided count")
			}},
	)
	assert.Nil(t, err, "FSM creation failed")

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

	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfConfig)

	// Setup 5GC
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.
		WithConfig(conf).
		WithFSM(f).
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
		securityContext.SetDefaultSNssai(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd})

		amfContext := fiveGC.GetAMFContext()
		amfContext.NewSecurityContext(securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 2000 ms
		time.Sleep(time.Duration(2000) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			assert.Equalf(t, state.Idle, ue.GetState().GetState().Current(), "Expected all ue to be in Idle state but was not")
		})
}
