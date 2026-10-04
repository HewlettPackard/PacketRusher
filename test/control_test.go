/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control"
	"my5G-RANTester/internal/control_test_engine/procedures"
	ngapConvert "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	ngap "github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
)

// startControlTest starts a UE on the first of two gNBs, and returns the control
// socket of the test. The mock core has no handover: it only acknowledges a Path
// Switch Request, and keeps finding the UE as both gNBs give it the same RAN UE
// NGAP ID.
func startControlTest(t *testing.T, n2, n3, amf string) (string, *context.Aio5gc) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort(n2), netip.MustParseAddrPort(n3),
		[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort(amf)}}})
	builder := aio5gc.FiveGCBuilder{}
	fiveGC, err := builder.WithConfig(conf).WithNGAPDispatcherHook(func(message ngap.Message, gnb *context.GNBContext, fiveGC *context.Aio5gc) (bool, error) {
		request, ok := message.(*ngap.PathSwitchRequest)
		if !ok {
			return false, nil
		}
		transfer, err := ngapConvert.Marshal(&ie.PathSwitchRequestAcknowledgeTransfer{ULNGUUPTNLInformation: ngapConvert.UPTransport(netip.MustParseAddr("127.0.0.1"), 71)})
		if err != nil {
			return true, err
		}
		acknowledge, err := (&ngap.PathSwitchRequestAcknowledge{
			AMFUENGAPID:     request.SourceAMFUENGAPID,
			RANUENGAPID:     request.RANUENGAPID,
			SecurityContext: &ie.SecurityContext{NextHopChainingCount: &ie.NextHopChainingCount{Value: 0}, NextHopNH: &ie.SecurityKey{Value: aper.BitString{Bytes: make([]byte, 32), BitLength: 256}}},
			AllowedNSSAI:    &ie.AllowedNSSAI{List: []ie.AllowedNSSAIItem{{SNSSAI: ngapConvert.Slice([]byte{1}, []byte{0, 0, 1})}}},
			PDUSessionResourceSwitchedList: &ie.PDUSessionResourceSwitchedList{List: []ie.PDUSessionResourceSwitchedItem{
				{PDUSessionID: &ie.PDUSessionID{Value: 1}, PathSwitchRequestAcknowledgeTransfer: ngapConvert.Octets(transfer)},
			}},
		}).MarshalBinary()
		if err == nil {
			gnb.SendMsg(acknowledge)
		}
		return true, err
	}).Build()
	require.NoError(t, err)
	t.Cleanup(func() { _ = fiveGC.Close() })

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(conf.Ue.Msin)
	securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})
	require.NoError(t, fiveGC.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext))

	wg := sync.WaitGroup{}
	gnbs, err := tools.CreateGnbs(t.Context(), 2, conf, &wg)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, gnb := range gnbs {
			gnb.Terminate()
		}
	})
	socket := filepath.Join(t.TempDir(), "control.sock")
	server, err := control.Listen(socket, gnbs, 1)
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Close() })
	simulation := tools.SimulateSingleUE(tools.UESimulationConfig{UeId: 1, Gnbs: gnbs, Cfg: conf, NumPduSessions: 2}, &wg)
	t.Cleanup(func() { stopTestSimulations(t, []*tools.UESimulation{simulation}) })
	server.AddUe(1, simulation)
	return socket, fiveGC
}

func registered(gnbId string) procedures.UeStatus {
	return procedures.UeStatus{UeId: 1, State: "registered", GnbId: gnbId, Connected: true, Ready: true, PduSessions: []int{1, 2}}
}

func TestControlSocket(t *testing.T) {
	socket, fiveGC := startControlTest(t, "127.0.0.1:9491", "127.0.0.1:2156", "127.0.0.1:38416")

	response, err := control.Call(socket, control.Request{Action: "inspect"})
	require.NoError(t, err)
	require.Len(t, response.Ues, 1)
	require.Equal(t, []control.Gnb{{Id: "000008", Ready: true}, {Id: "000009", Ready: true}}, response.Gnbs)

	response, err = control.Call(socket, control.Request{UeId: 1, Action: "wait"})
	require.NoError(t, err)
	require.Equal(t, registered("000008"), response.Ues[0])

	// Actions the UE is not in a state for are refused, and leave it as it was.
	for _, refused := range []control.Request{
		{UeId: 1, Action: "xn-handover", Target: "000008"},
		{UeId: 1, Action: "ng-handover", Target: "00000A"},
		{UeId: 1, Action: "reconnect"},
		{UeId: 1, Action: "register"},
		{UeId: 2, Action: "inspect"},
	} {
		_, err = control.Call(socket, refused)
		require.Error(t, err, refused)
	}

	response, err = control.Call(socket, control.Request{UeId: 1, Action: "xn-handover", Target: "000009"})
	require.NoError(t, err)
	require.Equal(t, registered("000009"), response.Ues[0])

	response, err = control.Call(socket, control.Request{UeId: 1, Action: "deregister"})
	require.NoError(t, err)
	require.Equal(t, "parked", response.Ues[0].State)
	require.Empty(t, response.Ues[0].PduSessions)
	require.Eventually(t, func() bool {
		_, err := fiveGC.GetAMFContext().FindRegisteredUEByMsin("0000000120")
		return err != nil
	}, 5*time.Second, 10*time.Millisecond, "the mock core must process the deregistration")

	// A parked UE has no registration to wait for until it is registered again.
	_, err = control.Call(socket, control.Request{UeId: 1, Action: "wait", TimeoutMs: 200})
	require.ErrorContains(t, err, "timed out")
}

func TestRunScenario(t *testing.T) {
	socket, _ := startControlTest(t, "127.0.0.1:9493", "127.0.0.1:2158", "127.0.0.1:38418")
	scenario := filepath.Join(t.TempDir(), "scenario.json")
	run := func(steps string) error {
		require.NoError(t, os.WriteFile(scenario, []byte(steps), 0600))
		return control.RunScenario(socket, scenario)
	}

	require.NoError(t, run(`{"steps": [
		{"ue": 1, "action": "wait", "timeout_ms": 30000},
		{"ue": 1, "action": "xn-handover", "target": "000009"},
		{"ue": 1, "action": "deregister"},
		{"ue": 1, "action": "register"},
		{"ue": 1, "action": "wait"}
	]}`))
	response, err := control.Call(socket, control.Request{UeId: 1, Action: "inspect"})
	require.NoError(t, err)
	require.Equal(t, registered("000008"), response.Ues[0])

	// A scenario stops at its first failing step, and runs none when one is invalid.
	require.ErrorContains(t, run(`{"steps": [{"ue": 1, "action": "reconnect"}, {"ue": 1, "action": "deregister"}]}`), "step 1")
	require.ErrorContains(t, run(`{"steps": [{"ue": 1, "action": "deregister"}, {"ue": 1, "action": "xn-handover"}]}`), "step 2")
	require.Error(t, run(`{"steps": [{"ue": 1, "action": "deregister", "timeout": 5}]}`))
	require.Error(t, run(`{"steps": []}`))
	response, err = control.Call(socket, control.Request{UeId: 1, Action: "inspect"})
	require.NoError(t, err)
	require.Equal(t, "registered", response.Ues[0].State)
}
