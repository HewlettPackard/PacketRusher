/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

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
		TimeBeforeHandover:       0,
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

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for 5 ms
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	time.Sleep(time.Duration(10000) * time.Millisecond)
	assert.Equalf(t, ueCount, createdSessionCount, "Expected %d PDU sessions created but was %d", ueCount, createdSessionCount)
}
