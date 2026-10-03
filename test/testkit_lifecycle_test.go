// SPDX-License-Identifier: Apache-2.0
package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/free5gc/ngap/message"
	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/service"
	"my5G-RANTester/test/aio5gc/testkit"
	"net/netip"
)

func TestNativeTestkitSimultaneousCoreIsolationAndTermination(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	firstCfg, secondCfg := testkit.LocalConfig(), testkit.LocalConfig()
	secondCfg.GNodeB.PlmnList.Mcc = "208"
	secondCfg.GNodeB.PlmnList.Mnc = "93"
	secondCfg.Ue.Hplmn.Mcc = "208"
	secondCfg.Ue.Hplmn.Mnc = "93"
	secondCfg.Ue.Key = "102132435465768798A9BACBDCEDFE0F"
	first, err := testkit.Start(ctx, new(aio5gc.FiveGCBuilder).WithConfig(firstCfg), 2)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, first.Close()) })
	second, err := testkit.Start(ctx, new(aio5gc.FiveGCBuilder).WithConfig(secondCfg), 1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, second.Close()) })
	require.NotEqual(t, first.Config.AMFs[0].AddrPort, second.Config.AMFs[0].AddrPort)
	require.NotZero(t, first.Config.AMFs[0].Port())
	require.Zero(t, firstCfg.AMFs[0].Port(), "builder must not rewrite caller's config")
	for _, fixture := range []*testkit.Fixture{first, second} {
		require.NoError(t, fixture.Provision(1))
		simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: 1, NumPduSessions: 1})
		require.NoError(t, err)
		attachment, err := simulation.Execute(ctx, "wait", "")
		require.NoError(t, err)
		require.Equal(t, []uint8{1}, attachment.ActivePDUSessions)
	}
	for _, fixture := range []*testkit.Fixture{first, second} {
		snapshot := fixture.Core.Snapshot()
		require.Len(t, snapshot.UEs, 1)
		require.Equal(t, int64(0), snapshot.UEs[0].AMFID)
		require.Equal(t, uint64(1), snapshot.UEs[0].Entries[core.Authenticated])
		require.Empty(t, snapshot.Errors)
		fixture.Core.GetAMFContext().ExecuteForAllUe(func(ue *core.UEContext) {
			require.True(t, strings.HasSuffix(ue.GetGuti(), "00000001"), "each independent core's first TMSI must be 1")
		})
		for _, association := range snapshot.Associations {
			local, err := netip.ParseAddrPort(association.Local)
			require.NoError(t, err)
			remote, err := netip.ParseAddrPort(association.Remote)
			require.NoError(t, err)
			require.Equal(t, fixture.Config.AMFs[0].AddrPort, local)
			require.Greater(t, remote.Port(), uint16(1024))
		}
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	require.NoError(t, first.CloseContext(shutdownCtx), "two gNBs and their signal waiters must join without process signals")
	require.Zero(t, first.Core.ResourceCount())
	require.Zero(t, first.Core.GetAMFContext().GNBCount())
	require.NoError(t, second.Provision(2))
	simulation, err := second.StartUE(tools.UESimulationConfig{UeId: 2, NumPduSessions: 1})
	require.NoError(t, err)
	_, err = simulation.Execute(ctx, "wait", "")
	require.NoError(t, err)
	require.Len(t, second.Core.Snapshot().UEs, 2)
	require.Empty(t, second.Core.Snapshot().Errors)
}
func TestNativeTestkitCloseCancelsBlockedDecodedPhase(t *testing.T) {
	gate := testkit.NewGate()
	builder := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig()).WithNGAPDispatcherHook(func(msg message.Message, gnb *core.GNBContext, fgc *core.Aio5gc) (bool, error) {
		if _, ok := msg.(*message.InitialUEMessage); ok {
			return true, gate.Wait(fgc.Context())
		}
		return false, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fixture, err := testkit.Start(ctx, builder, 2)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fixture.Close()) })
	require.NoError(t, fixture.Provision(1))
	simulation, err := fixture.StartUE(tools.UESimulationConfig{UeId: 1, NumPduSessions: 1})
	require.NoError(t, err)
	select {
	case <-gate.Reached():
	case <-ctx.Done():
		t.Fatal("native InitialUEMessage did not reach the gate")
	}
	require.Empty(t, fixture.Core.Snapshot().UEs, "decoded policy gate must run before normal registration")
	require.NoError(t, fixture.CloseContext(ctx), "shutdown must cancel the inline phase without manually opening it")
	select {
	case <-simulation.Done():
	default:
		t.Fatal("Close returned with an admitted UE still running")
	}
	require.Zero(t, fixture.Core.ResourceCount())
	require.Zero(t, fixture.Core.GetAMFContext().GNBCount())
}
func TestNativeEphemeralListenerCloseWakesAccept(t *testing.T) {
	listener, err := service.Listen(netip.MustParseAddrPort("127.0.0.1:0"))
	require.NoError(t, err)
	require.NotZero(t, listener.Addr().Port())
	var fgc core.Aio5gc
	fgc.RegisterCloser(listener.Close)
	require.True(t, fgc.Go(func() { service.Serve(listener, &fgc) }))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-listener.Accepting():
	case <-ctx.Done():
		t.Fatal("listener never waited in Accept")
	}
	require.NoError(t, fgc.CloseContext(ctx))
	require.NoError(t, listener.Close())
	// Closure released the actual port as well as joining the accept worker.
	replacement, err := service.Listen(listener.Addr())
	require.NoError(t, err)
	require.NoError(t, replacement.Close())
}

func TestNativeCoreObservesDispatchErrorsAndRetiresClosedPeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fgc, err := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig()).Build()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fgc.Close()) })
	addr, err := sctp.ResolveSCTPAddr("sctp", fgc.Config().AMFs[0].String())
	require.NoError(t, err)
	peer, err := sctp.DialSCTP("sctp", nil, addr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = peer.Close() })
	_, err = peer.Write([]byte{0xff})
	require.NoError(t, err)
	snapshot, err := fgc.Wait(ctx, func(s core.Snapshot) bool { return len(s.Errors) == 1 && len(s.Associations) == 1 })
	require.NoError(t, err)
	require.ErrorContains(t, snapshot.Errors[0], "NGAP decode")
	require.Equal(t, 2, fgc.ResourceCount())
	require.NoError(t, peer.Close())
	_, err = fgc.Wait(ctx, func(s core.Snapshot) bool { return len(s.Associations) == 0 })
	require.NoError(t, err)
	require.Equal(t, 1, fgc.ResourceCount(), "only listener remains owned after the reader exits")
}
func TestNativeBuilderBindFailureReleasesEarlierListener(t *testing.T) {
	cfg := testkit.LocalConfig()
	endpoint := netip.MustParseAddrPort("127.0.0.111:38588")
	cfg.AMFs = []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: endpoint}}, {IPv4Port: config.IPv4Port{AddrPort: endpoint}}}
	fgc, err := new(aio5gc.FiveGCBuilder).WithConfig(cfg).Build()
	require.Error(t, err)
	require.Nil(t, fgc)
	listener, err := service.Listen(endpoint)
	require.NoError(t, err, "failed multi-listener startup must release its first port")
	require.NoError(t, listener.Close())
}
