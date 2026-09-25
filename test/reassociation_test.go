/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */

package test

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
)

// The AMF drops the association while UEs are registered. The gNB must release the UE
// contexts that were served through it, re-establish the association with NG Setup, and
// accept new UEs on it -- rather than stop reading and leave the gNB serving nobody.
func TestGnbReestablishesLostAssociation(t *testing.T) {
	testReassociation(t, "127.0.0.1:9589", "127.0.0.1:2254", "127.0.0.1:38418", false)
}

// As above, but the AMF ignores the first NG Setup after the loss. The attempt must be
// abandoned after the setup timeout and retried, not left half-open.
func TestGnbRetriesUnansweredNgSetup(t *testing.T) {
	testReassociation(t, "127.0.0.1:9689", "127.0.0.1:2354", "127.0.0.1:38420", true)
}

// An AMF the gNB no longer serves through -- abandoned by InitGnb's startup retries, or
// removed by an AMF Configuration Update -- must not be re-dialled when its association
// closes, even though it was active.
func TestGnbDoesNotReestablishRemovedAmf(t *testing.T) {
	controlIFConfig := netip.MustParseAddrPort("127.0.0.1:9789")
	dataIFConfig := netip.MustParseAddrPort("127.0.0.1:2454")
	amfListConfig := []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38422")}},
	}
	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfListConfig)

	fiveGC, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).Build()
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	wg := sync.WaitGroup{}
	var gnb *gnbContext.GNBContext
	for _, g := range tools.CreateGnbs(1, conf, &wg) {
		gnb = g
	}
	var amf *gnbContext.GNBAmf
	for a := range gnb.IterGnbAmf() {
		if a.GetState() == gnbContext.Active {
			amf = a
		}
	}
	require.NotNil(t, amf, "the gNB should have completed NG Setup")

	gnb.DeleteGnBAmf(amf.GetAmfId())
	oldConn := amf.GetSCTPConn()
	amfSideGnb, err := fiveGC.GetAMFContext().GetGnb(oldConn.LocalAddr().String())
	require.NoError(t, err)
	require.NoError(t, amfSideGnb.GetSCTPConn().Close())

	require.Never(t, func() bool { return amf.GetSCTPConn() != oldConn }, 3*time.Second, 100*time.Millisecond,
		"a removed AMF should not be re-dialled")
}

func testReassociation(t *testing.T, n2, n3, amfAddr string, dropFirstSetup bool) {
	controlIFConfig := netip.MustParseAddrPort(n2)
	dataIFConfig := netip.MustParseAddrPort(n3)
	amfListConfig := []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort(amfAddr)}},
	}
	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfListConfig)

	// Armed just before the association is dropped, so the initial NG Setup is answered.
	var dropNextSetup atomic.Bool
	fiveGC, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).
		WithNGAPDispatcherHook(func(msg *ngapType.NGAPPDU, _ *context.GNBContext, _ *context.Aio5gc) (bool, error) {
			isSetup := msg.Present == ngapType.NGAPPDUPresentInitiatingMessage &&
				msg.InitiatingMessage.ProcedureCode.Value == ngapType.ProcedureCodeNGSetup
			return isSetup && dropNextSetup.CompareAndSwap(true, false), nil
		}).
		Build()
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(1, conf, &wg)
	require.Len(t, gnbs, 1)
	var gnb *gnbContext.GNBContext
	for _, g := range gnbs {
		gnb = g
	}
	var amf *gnbContext.GNBAmf
	for a := range gnb.IterGnbAmf() {
		if a.GetState() == gnbContext.Active {
			amf = a
		}
	}
	require.NotNil(t, amf, "the gNB should have completed NG Setup")

	registered := func() int {
		n := 0
		fiveGC.GetAMFContext().ExecuteForAllUe(func(ue *context.UEContext) {
			if ue.GetState().Current() == context.Registered {
				n++
			}
		})
		return n
	}
	uesOnAmf := func() int {
		n := 0
		gnb.GetUePool().Range(func(_, value any) bool {
			if value.(*gnbContext.GNBUe).GetAmfId() == amf.GetAmfId() {
				n++
			}
			return true
		})
		return n
	}
	simulate := func(ueId int) {
		cfg := tools.UESimulationConfig{Gnbs: gnbs, Cfg: conf, UeId: ueId}
		securityContext := context.SecurityContext{}
		securityContext.SetMsin(tools.IncrementMsin(ueId, conf.Ue.Msin))
		securityContext.SetAuthSubscription(conf.Ue.Key, conf.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
		securityContext.SetAbba([]uint8{0x00, 0x00})
		fiveGC.GetAMFContext().Provision(models.Snssai{Sst: int32(conf.Ue.Snssai.Sst), Sd: conf.Ue.Snssai.Sd}, securityContext)
		tools.SimulateSingleUE(cfg, &wg)
	}

	const ueCount = 3
	for id := 1; id <= ueCount; id++ {
		simulate(id)
	}
	require.Eventually(t, func() bool { return registered() == ueCount }, 10*time.Second, 100*time.Millisecond,
		"UEs should register before the association is lost")
	require.Equal(t, ueCount, uesOnAmf())

	plmnsBefore, slicesBefore := amf.GetLenPlmns(), amf.GetLenSlice()
	require.Positive(t, plmnsBefore)

	// The AMF side drops the association.
	dropNextSetup.Store(dropFirstSetup)
	lostAt := time.Now()
	oldConn := amf.GetSCTPConn()
	amfSideGnb, err := fiveGC.GetAMFContext().GetGnb(oldConn.LocalAddr().String())
	require.NoError(t, err)
	require.NoError(t, amfSideGnb.GetSCTPConn().Close())

	require.Eventually(t, func() bool {
		return amf.GetState() == gnbContext.Active && amf.GetSCTPConn() != oldConn
	}, 20*time.Second, 100*time.Millisecond, "the gNB should re-establish the association and complete NG Setup")
	if dropFirstSetup {
		require.False(t, dropNextSetup.Load(), "the first NG Setup after the loss should have been dropped")
		require.GreaterOrEqual(t, time.Since(lostAt), 5*time.Second,
			"recovery should have come from a retry after the setup timeout, not the first attempt")
	}
	require.Zero(t, uesOnAmf(), "UE contexts served through the lost association should be released")
	require.Equal(t, plmnsBefore, amf.GetLenPlmns(), "a second NG Setup should not append the same PLMNs again")
	require.Equal(t, slicesBefore, amf.GetLenSlice(), "a second NG Setup should not append the same slices again")

	// A UE arriving after recovery registers through the new association.
	simulate(ueCount + 1)
	require.Eventually(t, func() bool { return registered() == ueCount+1 }, 10*time.Second, 100*time.Millisecond,
		"a new UE should register through the re-established association")
	require.Equal(t, 1, uesOnAmf())
}
