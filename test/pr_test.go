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
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/util/fsm"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

	stateList := UEStateList{}
	f, err := fsm.NewFSM(
		fsm.Transitions{
			{Event: fsm.EventType(nas.MsgTypeRegistrationRequest), From: Fresh, To: RegistrationRequested},
			{Event: fsm.EventType(nas.MsgTypeAuthenticationResponse), From: RegistrationRequested, To: AuthenticationChallengeResponded},
			{Event: fsm.EventType(nas.MsgTypeSecurityModeComplete), From: AuthenticationChallengeResponded, To: SecurityContextSet},
			{Event: fsm.EventType(nas.MsgTypeRegistrationComplete), From: SecurityContextSet, To: Registred},
			{Event: fsm.EventType(nas.MsgTypeDeregistrationRequestUEOriginatingDeregistration), From: Registred, To: DeregistrationRequested},
			{Event: fsm.EventType(fmt.Sprint(ngapType.ProcedureCodeUEContextRelease)), From: DeregistrationRequested, To: Idle},
		},
		fsm.Callbacks{
			Fresh:                            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			RegistrationRequested:            func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			AuthenticationChallengeResponded: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			SecurityContextSet:               func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			Registred:                        func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			DeregistrationRequested:          func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {},
			Idle: func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
				assert.Equal(t, args["requestedPduCount"], args["providedPduCount"], "Number of requested PDU Session count does not match the provided count")
				assert.Equal(t, args["providedPduCount"], args["releasedPduCount"], "Number of provided PDU Session count does not match the released count")
			}},
	)
	assert.Nil(t, err, "FSM creation failed")
	stateList.Fsm = f

	logEntry := log.NewEntry(log.StandardLogger())

	validatePDUSessionCreation := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodePDUSessionResourceSetup {
				list := ngapMsg.SuccessfulOutcome.Value.PDUSessionResourceSetupResponse.ProtocolIEs.List
				for ie := range list {
					switch list[ie].Id.Value {
					case ngapType.ProtocolIEIDAMFUENGAPID:
						state, err := stateList.FindByAmfId(list[ie].Value.AMFUENGAPID.Value)
						assert.Nil(t, err, "Error while validating pdu session creation")
						state.NewPDUProvided()
					}
				}
			}
		}
		return false, nil
	}

	validatePDUSessionRelease := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodePDUSessionResourceRelease {
				list := ngapMsg.SuccessfulOutcome.Value.PDUSessionResourceReleaseResponse.ProtocolIEs.List
				for ie := range list {
					switch list[ie].Id.Value {
					case ngapType.ProtocolIEIDAMFUENGAPID:
						state, err := stateList.FindByAmfId(list[ie].Value.AMFUENGAPID.Value)
						assert.Nil(t, err, "Error while validating pdu session release")
						state.NewPDUReleased()
					}
				}
			}
		}
		return false, nil
	}

	validateContextRelease := func(ngapMsg *ngapType.NGAPPDU, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		if ngapMsg.Present == ngapType.NGAPPDUPresentSuccessfulOutcome {
			if ngapMsg.SuccessfulOutcome.ProcedureCode.Value == ngapType.ProcedureCodeUEContextRelease {
				list := ngapMsg.SuccessfulOutcome.Value.UEContextReleaseComplete.ProtocolIEs.List
				for ie := range list {
					switch list[ie].Id.Value {
					case ngapType.ProtocolIEIDAMFUENGAPID:
						ueState, err := stateList.FindByAmfId(list[ie].Value.AMFUENGAPID.Value)
						assert.Nil(t, err, "Error while validating pdu session release")
						err = stateList.Fsm.SendEvent(ueState.state, fsm.EventType(fmt.Sprint(ngapMsg.SuccessfulOutcome.ProcedureCode.Value)),
							fsm.ArgsType{
								"requestedPduCount": ueState.GetRequestedPDUCount(),
								"providedPduCount":  ueState.GetProvidedPDUCount(),
								"releasedPduCount":  ueState.GetReleasedPDUCount(),
							},
							logEntry)
						if err != nil {
							assert.Nil(t, err, "Wrong state transition: %v", err)
						}
					}
				}
			}
		}
		return false, nil
	}

	stateUpdate := func(nasMsg *nas.Message, ue *context.UEContext, gnb *context.GNBContext, fgc *context.Aio5gc) (bool, error) {
		var ueState *UEState
		var err error
		if nasMsg.GmmHeader.GetMessageType() == nas.MsgTypeRegistrationRequest {
			ueState, err = stateList.New(ue.GetAmfNgapId())
		} else {
			ueState, err = stateList.FindByAmfId(ue.GetAmfNgapId())
		}
		if err != nil {
			return false, err
		}

		switch nasMsg.GmmHeader.GetMessageType() {
		case nas.MsgTypeULNASTransport:
			ulNasTransport := nasMsg.ULNASTransport
			if ue.GetInitialContextSetup() && ulNasTransport.GetPayloadContainerType() == nasMessage.PayloadContainerTypeN1SMInfo {
				if ulNasTransport.RequestType != nil && ulNasTransport.RequestType.GetRequestTypeValue() == nasMessage.ULNASTransportRequestTypeInitialRequest {
					ueState.NewPDURequest()
				}
			}

		case nas.MsgTypeConfigurationUpdateComplete:
		default:
			err = stateList.Fsm.SendEvent(ueState.state, fsm.EventType(nasMsg.GmmHeader.GetMessageType()), fsm.ArgsType{}, logEntry)
			if err != nil {
				assert.Nil(t, err, "Wrong state transition: %v", err)
			}
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
		WithNGAPDispatcherHook(validateContextRelease).
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
		securityContext.SetDefaultSNssai(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd})

		amfContext := fiveGC.GetAMFContext()
		amfContext.NewSecurityContext(securityContext)

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	time.Sleep(time.Duration(5000) * time.Millisecond)
	uestates := stateList.UeStates
	for s := range uestates {
		assert.Equalf(t, Idle, uestates[s].state.Current(), "Expected all ue to be in Idle state but was not")
	}
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
		securityContext.SetDefaultSNssai(models.Snssai{Sst: int32(ueSimCfg.Cfg.Ue.Snssai.Sst), Sd: ueSimCfg.Cfg.Ue.Snssai.Sd})

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
