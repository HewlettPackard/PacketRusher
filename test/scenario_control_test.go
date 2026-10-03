// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/config"
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
	parked, err := sim.Execute(ctx, "deregister", "")
	require.NoError(t, err)
	require.Equal(t, "parked", parked.State)
	require.Eventually(t, func() bool { _, err := fgc.GetAMFContext().FindRegisteredUEByMsin(conf.Ue.Msin); return err != nil }, time.Second, 10*time.Millisecond)
	again, err := sim.Execute(ctx, "register", "")
	require.NoError(t, err)
	require.Equal(t, registered.Generation+1, again.Generation)
	require.True(t, again.Ready)
	require.Equal(t, []uint8{1}, again.ActivePDUSessions)
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
