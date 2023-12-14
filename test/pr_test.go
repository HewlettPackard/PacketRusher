/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"my5G-RANTester/test/state"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/nas"
	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

	stateList := state.UEStateList{}

	createdSessionCount := 0
	validatePDUSessionCreation := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodePDUSessionResourceSetup {
				createdSessionCount++
			}
		}
		return false, nil
	}

	releasedSessionCount := 0
	validatePDUSessionRelease := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodePDUSessionResourceRelease {
				releasedSessionCount++
			}
		}
		return false, nil
	}

	releasedCtxCount := 0
	validateCtxRelease := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodeUEContextRelease {
				releasedCtxCount++
			}
		}
		return false, nil
	}

	stateUpdate := func(nasMsg *nas.Message, ue *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		var ueState *state.UEState
		var err error
		if nasMsg.GmmHeader.GetMessageType() != nas.MsgTypeRegistrationRequest {
			ueState, err = stateList.FindByMsin(ue.GetSecurityContext().GetMsin())
			if err != nil {
				return false, err
			}
		}

		switch nasMsg.GmmHeader.GetMessageType() {
		// case nas.MsgTypeRegistrationRequest:
		// 	ueState.UpdateState(state.RegistrationRequested)

		case nas.MsgTypeAuthenticationResponse:
			ueState.UpdateState(state.AuthenticationChallengeResponded)

		case nas.MsgTypeSecurityModeComplete:
			ueState.UpdateState(state.SecurityContextSet)

		case nas.MsgTypeRegistrationComplete:
			ueState.UpdateState(state.Registred)

		case nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration:
			ueState.UpdateState(state.Idle)

		}
		if err != nil {
			log.Fatal(err)
		}

		return false, nil
	}

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
		WithNGAPDispatcherHook(validatePDUSessionCreation).
		WithNGAPDispatcherHook(validatePDUSessionRelease).
		WithNGAPDispatcherHook(validateCtxRelease).
		WithNASDispatcherHook(stateUpdate).
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
	ueCount := 1
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

		amfContext := fiveGC.GetAMFContext()
		amfContext.NewSecurityContext(securityContext)

		s := state.UEState{}
		expectedStates := []state.State{
			state.AuthenticationChallengeResponded,
			state.SecurityContextSet,
			state.Fresh,
			state.Registred,
			state.Idle,
		}
		s.Init(t, securityContext.GetMsin(), expectedStates, false)
		stateList.Add(&s)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	stateList.Check()
	assert.Equalf(t, ueCount, createdSessionCount, "Expected %d PDU sessions created but was %d", ueCount, createdSessionCount)
	assert.Equalf(t, ueCount, releasedSessionCount, "Expected %d PDU sessions created but was %d", ueCount, releasedSessionCount)
	assert.Equalf(t, ueCount, releasedCtxCount, "Expected %d PDU sessions created but was %d", ueCount, releasedCtxCount)
}

func TestUERegistrationLoop(t *testing.T) {

	completedRegistrationCount := 0
	validateRegistrationComplete := func(nasMsg *nas.Message, ue *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if nasMsg.GmmHeader.GetMessageType() == nas.MsgTypeRegistrationComplete {
			completedRegistrationCount++
		}
		return false, nil
	}

	releasedCtxCount := 0
	validateCtxRelease := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodeUEContextRelease {
				releasedCtxCount++
			}
		}
		return false, nil
	}

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
		WithNASDispatcherHook(validateRegistrationComplete).
		WithNGAPDispatcherHook(validateCtxRelease).
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
		amfContext.NewSecurityContext(securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 2000 ms
		time.Sleep(time.Duration(2000) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	assert.Equalf(t, loopCount, completedRegistrationCount, "Expected %d completed registrations but was %d", loopCount, completedRegistrationCount)
	assert.Equalf(t, loopCount, releasedCtxCount, "Expected %d contexts releases but was %d", loopCount, releasedCtxCount)
}
