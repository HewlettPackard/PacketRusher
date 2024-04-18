/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/scenario"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	testCount = 0
	mutex     sync.Mutex
)

func TestSingleUe(t *testing.T) {
	conf := generateDefaultConf()
	type UECheck struct {
		HasAuthOnce bool
	}
	ueChecks := map[string]*UECheck{}

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

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(conf.Ue.Msin)
	securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})

	amfContext := fiveGC.GetAMFContext()
	amfContext.Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext)

	gnbs := []config.GNodeB{conf.GNodeB}
	ueScenario := scenario.UEScenario{
		Config: conf.Ue,
		Loop:   false,
		Hang:   false,
		Tasks: []scenario.Task{
			{
				TaskType: scenario.AttachToGNB,
				Parameters: struct {
					GnbId string
				}{conf.GNodeB.PlmnList.GnbId},
			},
			{
				TaskType: scenario.Registration,
			},
			{
				TaskType: scenario.NewPDUSession,
			},
			{
				TaskType: scenario.Deregistration,
				Delay:    2000,
			},
		},
	}
	ueScenarios := []scenario.UEScenario{ueScenario}

	scenario.Start(gnbs, conf.AMFs, ueScenarios, 50, 30000)

	time.Sleep(2 * time.Second)

	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			assert.True(t, ueChecks[ue.GetSecurityContext().GetMsin()].HasAuthOnce, "UE has never changed state")
		})
}

func TestRegistrationToCtxReleaseWithPDUSession(t *testing.T) {

	conf := generateDefaultConf()

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

	// Setup UE
	ueCount := 10
	amfContext := fiveGC.GetAMFContext()
	ueScenarios := []scenario.UEScenario{}
	for i := 1; i <= ueCount; i++ {
		tmpConf := conf
		tmpConf.Ue.Msin = tools.IncrementMsin(i, conf.Ue.Msin)
		ueScenario := scenario.UEScenario{
			Config: tmpConf.Ue,
			Loop:   false,
			Hang:   false,
			Tasks: []scenario.Task{
				{
					TaskType: scenario.AttachToGNB,
					Parameters: struct {
						GnbId string
					}{tmpConf.GNodeB.PlmnList.GnbId},
				},
				{
					TaskType: scenario.Registration,
				},
				{
					TaskType: scenario.NewPDUSession,
				},
				{
					TaskType: scenario.Deregistration,
					Delay:    2000,
				},
			},
		}
		ueScenarios = append(ueScenarios, ueScenario)

		securityContext := context.SecurityContext{}
		securityContext.SetMsin(tmpConf.Ue.Msin)
		securityContext.SetAuthSubscription(tmpConf.Ue.Key, tmpConf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})

		amfContext.Provision(models.Snssai{Sst: int32(tmpConf.Ue.Snssai.Sst), Sd: tmpConf.Ue.Snssai.Sd}, securityContext)

	}

	gnbs := []config.GNodeB{conf.GNodeB}
	scenario.Start(gnbs, conf.AMFs, ueScenarios, 500, 40000)

	time.Sleep(2 * time.Second)

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

	type UECheck struct {
		HasAuthOnce bool
	}
	ueChecks := map[string]*UECheck{}

	conf := generateDefaultConf()

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

	// Setup UE
	securityContext := context.SecurityContext{}
	securityContext.SetMsin(conf.Ue.Msin)
	securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})

	amfContext := fiveGC.GetAMFContext()
	amfContext.Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext)

	gnbs := []config.GNodeB{conf.GNodeB}
	ueScenario := scenario.UEScenario{
		Config: conf.Ue,
		Loop:   true,
		Hang:   false,
		Tasks: []scenario.Task{
			{
				TaskType: scenario.AttachToGNB,
				Parameters: struct {
					GnbId string
				}{conf.GNodeB.PlmnList.GnbId},
			},
			{
				TaskType: scenario.Registration,
			},
			{
				TaskType: scenario.NewPDUSession,
			},
			{
				TaskType: scenario.Deregistration,
				Delay:    2000,
			},
		},
	}
	ueScenarios := []scenario.UEScenario{ueScenario}

	scenario.Start(gnbs, conf.AMFs, ueScenarios, 50, 15000)

	time.Sleep(2 * time.Second)
	i := 0
	fiveGC.GetAMFContext().ExecuteForAllUe(
		func(ue *context.UEContext) {
			i++
			assert.Equalf(t, context.Deregistered, ue.GetState().Current(), "Expected all ue to be in Deregistered state but was not")
			assert.True(t, ueChecks[ue.GetSecurityContext().GetMsin()].HasAuthOnce, "UE has never changed state")

		})
	assert.GreaterOrEqual(t, i, 3)
}

func generateDefaultConf() config.Config {

	conf := config.Config{
		GNodeB: config.GNodeB{
			ControlIF: config.ControlIF{
				Ip:   "127.0.0.1",
				Port: 9489,
			},
			DataIF: config.DataIF{
				Ip:   "127.0.0.1",
				Port: 2154,
			},
			PlmnList: config.PlmnList{
				Mcc:   "999",
				Mnc:   "70",
				Tac:   "000001",
				GnbId: "000008",
			},
			SliceSupportList: config.SliceSupportList{
				Sst: "01",
				Sd:  "000001",
			},
		},
		Ue: config.Ue{
			Msin:             "0000000120",
			Key:              "00112233445566778899AABBCCDDEEFF",
			Opc:              "00112233445566778899AABBCCDDEEFF",
			Amf:              "8000",
			Sqn:              "00000000",
			Dnn:              "internet",
			RoutingIndicator: "4567",
			Hplmn: config.Hplmn{
				Mcc: "999",
				Mnc: "70",
			},
			Snssai: config.Snssai{
				Sst: 01,
				Sd:  "000001",
			},
			Integrity: config.Integrity{
				Nia0: false,
				Nia1: true,
				Nia2: true,
			},
			Ciphering: config.Ciphering{
				Nea0: false,
				Nea1: true,
				Nea2: true,
			},
		},
		AMFs: []*config.AMF{
			{
				Ip:   "127.0.0.1",
				Port: 38414,
			}},
		Logs: config.Logs{
			Level: 4,
		},
	}

	mutex.Lock()
	conf.AMFs[0].Port = conf.AMFs[0].Port + testCount
	conf.GNodeB.ControlIF.Port = conf.GNodeB.ControlIF.Port + testCount
	conf.GNodeB.DataIF.Port = conf.GNodeB.DataIF.Port + testCount
	testCount++
	mutex.Unlock()
	return conf
}
