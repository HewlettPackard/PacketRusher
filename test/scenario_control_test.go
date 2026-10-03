// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/scenario"
	ngapConvert "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	coreConvert "my5G-RANTester/test/aio5gc/lib/convert"
	coreTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	ngap "github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/stretchr/testify/require"
)

func TestScenarioControlsWaitForPDUAndTargetedXnCompletion(t *testing.T) {
	conf := coreTools.GenerateDefaultConf(netip.AddrPortFrom(netip.MustParseAddr("127.0.0.31"), uint16(30000+os.Getpid()%10000)), netip.MustParseAddrPort("127.0.0.31:2161"), []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(40000+os.Getpid()%10000))}}})
	pathSwitch := make(chan *ngap.PathSwitchRequest, 1)
	allowAck := make(chan struct{})
	var unblock sync.Once
	release := func() { unblock.Do(func() { close(allowAck) }) }
	t.Cleanup(release)
	handoverRequired := make(chan *ngap.HandoverRequired, 2)
	handoverNotify := make(chan *ngap.HandoverNotify, 1)
	var answerNG atomic.Bool
	var coreMu sync.Mutex
	coreGnbs := make(map[string]*core.GNBContext)
	var handoverSource *core.GNBContext
	var handoverSourceRan *ie.RANUENGAPID
	builder := aio5gc.FiveGCBuilder{}
	fgc, err := builder.WithConfig(conf).WithNGAPDispatcherHook(func(message ngap.Message, gnb *core.GNBContext, fgc *core.Aio5gc) (bool, error) {
		switch request := message.(type) {
		case *ngap.NGSetupRequest:
			id := request.GlobalRANNodeID.Choice.(*ie.GlobalGNBID).GNBID.Choice.(*ie.GNBIDForGNBID)
			coreMu.Lock()
			coreGnbs[fmt.Sprintf("%x", id.Value.Bytes)] = gnb
			coreMu.Unlock()
		case *ngap.PathSwitchRequest:
			pathSwitch <- request
			<-allowAck
			ue, err := fgc.GetAMFContext().FindUEById(request.SourceAMFUENGAPID.Value)
			if err != nil {
				return true, err
			}
			ue.SetRanNgapId(request.RANUENGAPID.Value)
			transfer, err := ngapConvert.Marshal(&ie.PathSwitchRequestAcknowledgeTransfer{ULNGUUPTNLInformation: ngapConvert.UPTransport(netip.MustParseAddr("127.0.0.1"), 71)})
			if err != nil {
				return true, err
			}
			ack := &ngap.PathSwitchRequestAcknowledge{AMFUENGAPID: request.SourceAMFUENGAPID, RANUENGAPID: request.RANUENGAPID,
				SecurityContext:                &ie.SecurityContext{NextHopChainingCount: &ie.NextHopChainingCount{Value: 0}, NextHopNH: &ie.SecurityKey{Value: aper.BitString{Bytes: make([]byte, 32), BitLength: 256}}},
				AllowedNSSAI:                   &ie.AllowedNSSAI{List: []ie.AllowedNSSAIItem{{SNSSAI: &ie.SNSSAI{SST: &ie.SST{Value: []byte{1}}, SD: &ie.SD{Value: []byte{0, 0, 1}}}}}},
				PDUSessionResourceSwitchedList: &ie.PDUSessionResourceSwitchedList{List: []ie.PDUSessionResourceSwitchedItem{{PDUSessionID: &ie.PDUSessionID{Value: 1}, PathSwitchRequestAcknowledgeTransfer: ngapConvert.Octets(transfer)}}},
			}
			wire, err := ack.MarshalBinary()
			if err == nil {
				gnb.SendMsg(wire)
			}
			return true, err
		case *ngap.HandoverRequired:
			handoverRequired <- request
			if !answerNG.Load() {
				return true, nil
			} // The first action must reach its finite deadline.
			id := request.TargetID.Choice.(*ie.TargetRANNodeID).GlobalRANNodeID.Choice.(*ie.GlobalGNBID).GNBID.Choice.(*ie.GNBIDForGNBID)
			coreMu.Lock()
			target := coreGnbs[fmt.Sprintf("%x", id.Value.Bytes)]
			handoverSource, handoverSourceRan = gnb, request.RANUENGAPID
			coreMu.Unlock()
			if target == nil {
				return true, fmt.Errorf("unknown target gNB")
			}
			wire, err := scenarioHandoverRequest(request, fgc)
			if err == nil {
				target.SendMsg(wire)
			}
			return true, err
		case *ngap.HandoverRequestAcknowledge:
			coreMu.Lock()
			source, ran := handoverSource, handoverSourceRan
			coreMu.Unlock()
			if source == nil {
				return true, fmt.Errorf("handover ACK without source")
			}
			wire, err := (&ngap.HandoverCommand{AMFUENGAPID: request.AMFUENGAPID, RANUENGAPID: ran,
				HandoverType: &ie.HandoverType{Value: ie.HandoverTypePresentIntra5gs}, TargetToSourceTransparentContainer: request.TargetToSourceTransparentContainer}).MarshalBinary()
			if err == nil {
				source.SendMsg(wire)
			}
			return true, err
		case *ngap.HandoverNotify:
			ue, err := fgc.GetAMFContext().FindUEById(request.AMFUENGAPID.Value)
			if err == nil {
				ue.SetRanNgapId(request.RANUENGAPID.Value)
			}
			handoverNotify <- request
			return true, err
		}
		return false, nil
	}).Build()
	require.NoError(t, err)
	t.Cleanup(func() { release(); require.NoError(t, fgc.Close()) })
	var wg sync.WaitGroup
	gnbs := tools.CreateGnbs(2, conf, &wg)
	t.Cleanup(func() {
		for _, node := range gnbs {
			node.Terminate()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	require.NoError(t, tools.WaitGnbs(ctx, gnbs))
	registry := scenario.NewRegistry(gnbs)
	socket := filepath.Join(t.TempDir(), "control.sock")
	server, err := scenario.Listen(socket, registry)
	require.NoError(t, err)
	t.Cleanup(func() { server.Close() })
	var simulations []*tools.UESimulation
	t.Cleanup(func() {
		for _, sim := range simulations {
			sim.Send(procedures.UeTesterMessage{Type: procedures.Kill})
			<-sim.Done()
		}
	})
	for id := 1; id <= 2; id++ {
		security := core.SecurityContext{}
		security.SetMsin(tools.IncrementMsin(id, conf.Ue.Msin))
		security.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		security.SetAbba([]byte{0, 0})
		require.NoError(t, fgc.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, security))
		sim := tools.SimulateSingleUE(tools.UESimulationConfig{UeId: id, Cfg: conf, Gnbs: gnbs, NumPduSessions: 1}, &wg)
		simulations = append(simulations, sim)
		registry.Add(id, sim)
		response, err := scenario.Call(ctx, socket, scenario.Request{Action: "wait", UE: id, TimeoutMS: 15000})
		require.NoError(t, err)
		require.True(t, response.UEs[0].Ready)
		require.Equal(t, []uint8{1}, response.UEs[0].ActivePDUSessions)
	}
	finished := make(chan error, 1)
	go func() {
		_, err := scenario.Call(ctx, socket, scenario.Request{UE: 1, Action: "xn-handover", Target: "000009", TimeoutMS: 15000})
		finished <- err
	}()
	select {
	case request := <-pathSwitch:
		require.NotNil(t, request.PDUSessionResourceToBeSwitchedDLList)
	case <-ctx.Done():
		t.Fatal("target did not send Path Switch")
	}
	select {
	case err := <-finished:
		t.Fatalf("handover completed before core acknowledgement: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	response, err := scenario.Call(ctx, socket, scenario.Request{Action: "inspect", UE: 2})
	require.NoError(t, err)
	require.Equal(t, "000009", response.UEs[0].GNB)
	require.True(t, response.UEs[0].Ready)
	release()
	select {
	case err := <-finished:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("handover did not complete")
	}
	response, err = scenario.Call(ctx, socket, scenario.Request{Action: "inspect", UE: 1})
	require.NoError(t, err)
	require.Equal(t, "000009", response.UEs[0].GNB)
	require.True(t, response.UEs[0].Ready)
	response, err = scenario.Call(ctx, socket, scenario.Request{Action: "ng-handover", UE: 1, Target: "000008", TimeoutMS: 200})
	require.Error(t, err)
	select {
	case request := <-handoverRequired:
		target := request.TargetID.Choice.(*ie.TargetRANNodeID).GlobalRANNodeID.Choice.(*ie.GlobalGNBID).GNBID.Choice.(*ie.GNBIDForGNBID)
		require.Equal(t, []byte{0, 0, 8}, target.Value.Bytes)
	case <-time.After(time.Second):
		t.Fatal("NG handover did not reach the core")
	}
	require.Eventually(t, func() bool {
		a, err := simulations[0].Inspect(ctx)
		return err == nil && a.Ready && a.GNB == "000009"
	}, time.Second, 10*time.Millisecond, "timed-out NG handover left its source permanently pending")
	answerNG.Store(true)
	response, err = scenario.Call(ctx, socket, scenario.Request{Action: "ng-handover", UE: 1, Target: "000008", TimeoutMS: 15000})
	require.NoError(t, err)
	require.Equal(t, "000008", response.UEs[0].GNB)
	require.True(t, response.UEs[0].Ready)
	require.Equal(t, []uint8{1}, response.UEs[0].ActivePDUSessions)
	select {
	case <-handoverNotify:
	case <-ctx.Done():
		t.Fatal("completed NG handover did not notify the core")
	}
	// Cancellation must not queue an action that executes during a later generation.
	cancelled, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()
	_, err = simulations[1].Execute(cancelled, "deregister", "")
	require.True(t, errors.Is(err, context.Canceled))
}

func TestScenarioDeregistrationParksAndRegistrationRearms(t *testing.T) {
	conf := coreTools.GenerateDefaultConf(netip.AddrPortFrom(netip.MustParseAddr("127.0.0.41"), uint16(30000+os.Getpid()%10000)), netip.MustParseAddrPort("127.0.0.41:2162"), []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(50000+os.Getpid()%10000))}}})
	builder := aio5gc.FiveGCBuilder{}
	fgc, err := builder.WithConfig(conf).Build()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fgc.Close()) })
	security := core.SecurityContext{}
	security.SetMsin(conf.Ue.Msin)
	security.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	security.SetAbba([]byte{0, 0})
	require.NoError(t, fgc.GetAMFContext().Provision(models.Snssai{Sst: 1, Sd: conf.Ue.Snssai.Sd}, security))
	var wg sync.WaitGroup
	gnbs := tools.CreateGnbs(1, conf, &wg)
	t.Cleanup(func() {
		for _, node := range gnbs {
			node.Terminate()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	require.NoError(t, tools.WaitGnbs(ctx, gnbs))
	sim := tools.SimulateSingleUE(tools.UESimulationConfig{UeId: 1, Cfg: conf, Gnbs: gnbs, NumPduSessions: 1}, &wg)
	t.Cleanup(func() { sim.Send(procedures.UeTesterMessage{Type: procedures.Kill}); <-sim.Done() })
	registered, err := sim.Execute(ctx, "wait", "")
	require.NoError(t, err)
	coreUE, err := fgc.GetAMFContext().FindRegisteredUEByMsin(conf.Ue.Msin)
	require.NoError(t, err)
	parked, err := sim.Execute(ctx, "deregister", "")
	require.NoError(t, err)
	require.Equal(t, "parked", parked.State)
	// Switch-off has no NAS acknowledgement. Observe the original core UE's
	// deregistered/inactive state before rearming, within a finite procedure bound.
	require.Eventually(t, func() bool {
		if coreUE.GetState().Current() != core.Deregistered {
			return false
		}
		for _, session := range coreUE.GetSmContexts() {
			if session.GetState().Current() != core.Inactive {
				return false
			}
		}
		_, err := fgc.GetAMFContext().FindRegisteredUEByMsin(conf.Ue.Msin)
		return err != nil
	}, 10*time.Second, 10*time.Millisecond)
	again, err := sim.Execute(ctx, "register", "")
	require.NoError(t, err)
	require.Equal(t, registered.Generation+1, again.Generation)
	require.True(t, again.Ready)
	require.Equal(t, []uint8{1}, again.ActivePDUSessions)
}

// Hold the core before it emits the encoded second PDU accept. The first
// session is already active; automatic td must wait for both configured IDs.
func TestScenarioAutomaticDeregistrationWaitsForEveryPDUAccept(t *testing.T) {
	results := analytics.NewRecorder()
	analytics.SetCurrent(results)
	t.Cleanup(func() { analytics.SetCurrent(nil) })
	conf := coreTools.GenerateDefaultConf(netip.AddrPortFrom(netip.MustParseAddr("127.0.0.51"), uint16(27000+os.Getpid()%2000)), netip.MustParseAddrPort("127.0.0.51:2163"), []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(60000+os.Getpid()%4000))}}})
	pendingAccept, allowAccept := make(chan struct{}), make(chan struct{})
	var block, unblock sync.Once
	release := func() { unblock.Do(func() { close(allowAccept) }) }
	builder := aio5gc.FiveGCBuilder{}
	fgc, err := builder.WithConfig(conf).WithPDUCallback(core.Active, func(state *fsm.State, event fsm.EventType, args fsm.ArgsType) {
		if event == fsm.EntryEvent && args["sm"].(*core.SmContext).GetPduSessionId() == 2 {
			block.Do(func() { close(pendingAccept) })
			<-allowAccept
		}
	}).Build()
	require.NoError(t, err)
	t.Cleanup(func() { release(); require.NoError(t, fgc.Close()) })
	security := core.SecurityContext{}
	security.SetMsin(conf.Ue.Msin)
	security.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	security.SetAbba([]byte{0, 0})
	require.NoError(t, fgc.GetAMFContext().Provision(models.Snssai{Sst: 1, Sd: conf.Ue.Snssai.Sd}, security))
	var wg sync.WaitGroup
	gnbs := tools.CreateGnbs(1, conf, &wg)
	t.Cleanup(func() {
		for _, node := range gnbs {
			node.Terminate()
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, tools.WaitGnbs(ctx, gnbs))
	sim := tools.SimulateSingleUE(tools.UESimulationConfig{UeId: 1, Cfg: conf, Gnbs: gnbs, NumPduSessions: 2, TimeBeforeDeregistration: 400}, &wg)
	t.Cleanup(func() {
		release()
		sim.Send(procedures.UeTesterMessage{Type: procedures.Kill})
		select {
		case <-sim.Done():
		case <-time.After(5 * time.Second):
			t.Error("simulation did not shut down")
		}
	})
	select {
	case <-pendingAccept:
	case <-ctx.Done():
		t.Fatal("second PDU request did not reach the core")
	}
	require.Eventually(t, func() bool {
		a, err := sim.Inspect(ctx)
		return err == nil && a.State == "registered" && len(a.ActivePDUSessions) == 1 && a.ActivePDUSessions[0] == 1 && !a.Ready
	}, 5*time.Second, 10*time.Millisecond)
	// This deliberate response delay is longer than td400ms, rather than a
	// readiness sleep: termination must remain absent throughout the interval.
	select {
	case <-sim.Done():
		t.Fatal("automatic deregistration cancelled a pending PDU establishment")
	case <-time.After(700 * time.Millisecond):
	}
	a, err := sim.Inspect(ctx)
	require.NoError(t, err, "inspection must remain available during automatic readiness wait")
	require.Equal(t, []uint8{1}, a.ActivePDUSessions)
	require.False(t, a.Ready)
	release()
	select {
	case <-sim.Done():
	case <-ctx.Done():
		t.Fatal("ready UE did not complete automatic deregistration")
	}
	require.Eventually(t, func() bool {
		seen, complete := 0, true
		fgc.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) {
			seen++
			complete = complete && ue.GetState().Current() == core.Deregistered
			ue.ExecuteForAllSmContexts(func(sm *core.SmContext) {
				complete = complete && sm.GetState().Current() == core.Inactive
			})
		})
		return seen == 1 && complete
	}, 5*time.Second, 20*time.Millisecond)
	for _, procedure := range results.Snapshot().Procedures {
		want := uint64(1)
		if procedure.Procedure == analytics.SessionEstablishment {
			want = 2
		}
		require.Equal(t, want, procedure.Started)
		require.Equal(t, want, procedure.Success)
		require.Zero(t, procedure.Failure)
		require.Zero(t, procedure.Cancelled)
		require.Zero(t, procedure.Pending)
	}
}

// Complete the core side over encoded NGAP; production gNB handlers create the
// target context, acknowledge it, move the UE connection, and emit Handover Notify.
func scenarioHandoverRequest(required *ngap.HandoverRequired, fgc *core.Aio5gc) ([]byte, error) {
	transfer, err := ngapConvert.Marshal(&ie.PDUSessionResourceSetupRequestTransfer{ProtocolIEs: &ie.ProtocolIEContainerPDUSessionResourceSetupRequestTransferIEs{List: []ie.PDUSessionResourceSetupRequestTransferIEs{{ULNGUUPTNLInformation: ngapConvert.UPTransport(netip.MustParseAddr("127.0.0.1"), 71)}}}})
	if err != nil {
		return nil, err
	}
	slice := &ie.SNSSAI{SST: &ie.SST{Value: []byte{1}}, SD: &ie.SD{Value: []byte{0, 0, 1}}}
	request := &ngap.HandoverRequest{AMFUENGAPID: required.AMFUENGAPID, HandoverType: required.HandoverType, Cause: required.Cause,
		UEAggregateMaximumBitRate: &ie.UEAggregateMaximumBitRate{UEAggregateMaximumBitRateDL: &ie.BitRate{Value: 1000000}, UEAggregateMaximumBitRateUL: &ie.BitRate{Value: 1000000}},
		UESecurityCapabilities: &ie.UESecurityCapabilities{
			NRencryptionAlgorithms:             &ie.NRencryptionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
			NRintegrityProtectionAlgorithms:    &ie.NRintegrityProtectionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
			EUTRAencryptionAlgorithms:          &ie.EUTRAencryptionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
			EUTRAintegrityProtectionAlgorithms: &ie.EUTRAintegrityProtectionAlgorithms{Value: aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}},
		},
		SecurityContext:                  &ie.SecurityContext{NextHopChainingCount: &ie.NextHopChainingCount{Value: 0}, NextHopNH: &ie.SecurityKey{Value: aper.BitString{Bytes: make([]byte, 32), BitLength: 256}}},
		PDUSessionResourceSetupListHOReq: &ie.PDUSessionResourceSetupListHOReq{List: []ie.PDUSessionResourceSetupItemHOReq{{PDUSessionID: &ie.PDUSessionID{Value: 1}, SNSSAI: slice, HandoverRequestTransfer: ngapConvert.Octets(transfer)}}},
		AllowedNSSAI:                     &ie.AllowedNSSAI{List: []ie.AllowedNSSAIItem{{SNSSAI: slice}}}, SourceToTargetTransparentContainer: required.SourceToTargetTransparentContainer,
		GUAMI: coreConvert.GUAMIToNGAP(fgc.GetAMFContext().GetServedGuami()[0]),
	}
	return request.MarshalBinary()
}
