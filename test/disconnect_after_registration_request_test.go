/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/tools"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With DisconnectAfterRegistrationRequest, each UE sends its Registration Request,
// then drops its radio link: the core receives every request, no registration
// completes, and the scenarios end on their own, once per loop iteration.
func TestDisconnectAfterRegistrationRequest(t *testing.T) {
	for _, tc := range []struct {
		name      string
		loopCount int
		gnbPort   uint16
		dataPort  uint16
		amfPort   uint16
	}{
		{name: "single registration", loopCount: 0, gnbPort: 9590, dataPort: 2754, amfPort: 38428},
		{name: "registration loop", loopCount: 3, gnbPort: 9591, dataPort: 2755, amfPort: 38430},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := analytics.Results()

			conf := amfTools.GenerateDefaultConf(
				netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), tc.gnbPort),
				netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), tc.dataPort),
				[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), tc.amfPort)}}},
			)

			var requestsMu sync.Mutex
			// The core starts authenticating a UE before it has resolved its identity.
			requests := 0
			authenticated := 0
			fiveGC, err := (&aio5gc.FiveGCBuilder{}).
				WithConfig(conf).
				WithUeCallback(context.AuthenticationInitiated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
					if event != fsm.EntryEvent {
						return
					}
					requestsMu.Lock()
					defer requestsMu.Unlock()
					requests++
				}).
				WithUeCallback(context.Authenticated, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
					if event != fsm.EntryEvent {
						return
					}
					requestsMu.Lock()
					defer requestsMu.Unlock()
					authenticated++
				}).
				Build()
			require.NoError(t, err)
			t.Cleanup(func() { _ = fiveGC.Close() })

			wg := sync.WaitGroup{}
			gnbs, err := tools.CreateGnbs(t.Context(), 1, conf, &wg)
			require.NoError(t, err)
			t.Cleanup(func() {
				for _, gnb := range gnbs {
					gnb.Terminate()
				}
			})

			ueCount := 5
			ueSimCfg := tools.UESimulationConfig{
				Gnbs:                               gnbs,
				Cfg:                                conf,
				NumPduSessions:                     1,
				RegistrationLoop:                   tc.loopCount != 0,
				LoopCount:                          tc.loopCount,
				TimeBeforeReregistration:           10,
				DisconnectAfterRegistrationRequest: true,
			}
			simulations := make([]*tools.UESimulation, 0, ueCount)
			t.Cleanup(func() { stopTestSimulations(t, simulations) })
			for ueSimCfg.UeId = 1; ueSimCfg.UeId <= ueCount; ueSimCfg.UeId++ {
				securityContext := context.SecurityContext{}
				securityContext.SetMsin(tools.IncrementMsin(ueSimCfg.UeId, conf.Ue.Msin))
				securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
				securityContext.SetAbba([]uint8{0x00, 0x00})
				fiveGC.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext)

				simulations = append(simulations, tools.SimulateSingleUE(ueSimCfg, &wg))
				time.Sleep(5 * time.Millisecond)
			}

			// Every scenario ends without being told to: its UEs disconnected.
			waitTestSimulations(t, simulations, 30*time.Second)

			iterations := max(tc.loopCount, 1)
			assert.Equal(t, uint64(ueCount*iterations), analytics.Results()[analytics.Registration].Started-before[analytics.Registration].Started,
				"every iteration of every UE must send a Registration Request")
			assert.Equal(t, before[analytics.Registration].Success, analytics.Results()[analytics.Registration].Success,
				"no registration may complete")
			assert.Equal(t, before[analytics.SessionEstablishment].Started, analytics.Results()[analytics.SessionEstablishment].Started,
				"no PDU session may be requested")

			// The core received every Registration Request, each on a new UE context
			// as the UE disconnected before identifying itself further. It answers it
			// with an Authentication Request, which no UE is left to answer.
			require.Eventually(t, func() bool {
				requestsMu.Lock()
				defer requestsMu.Unlock()
				coreUes := 0
				fiveGC.GetAMFContext().ExecuteForAllUe(func(*context.UEContext) { coreUes++ })
				return requests == ueCount*iterations && coreUes == ueCount*iterations
			}, 5*time.Second, 10*time.Millisecond, "the core must receive every Registration Request")
			fiveGC.GetAMFContext().ExecuteForAllUe(func(ue *context.UEContext) {
				assert.NotEqual(t, context.Registered, ue.GetState().Current(), "no UE may register")
			})
			requestsMu.Lock()
			assert.Zero(t, authenticated, "no UE may answer the authentication")
			requestsMu.Unlock()

			// The gNB took every UE's radio link down.
			for _, gnb := range gnbs {
				require.Eventually(t, func() bool {
					up := 0
					gnb.GetUePool().Range(func(_, value any) bool {
						if value.(*gnbContext.GNBUe).GetState() != gnbContext.Down {
							up++
						}
						return true
					})
					return up == 0
				}, 5*time.Second, 10*time.Millisecond, "the gNB must see every UE disconnect")
			}
		})
	}
}
