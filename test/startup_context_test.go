// SPDX-License-Identifier: Apache-2.0
package test

import (
	"bytes"
	"context"
	"my5G-RANTester/internal/common/tools"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/testkit"
	"net/netip"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
	"my5G-RANTester/test/aio5gc/service"
)

// The second real encoded NG Setup is blocked after the first gNB succeeded.
// Cancellation must own partial construction and join both signal/readers,
// rather than waiting through the legacy five startup retries.
func TestNativeStartupContextCancelsPartialConstruction(t *testing.T) {
	gate := testkit.NewGate()
	observed := make(chan *core.Aio5gc, 1)
	builder := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig())
	var reached sync.Once
	builder.WithNGAPDispatcherHook(func(msg message.Message, _ *core.GNBContext, fgc *core.Aio5gc) (bool, error) {
		setup, ok := msg.(*message.NGSetupRequest)
		if !ok {
			return false, nil
		}
		node := setup.GlobalRANNodeID.Choice.(*ie.GlobalGNBID)
		id := node.GNBID.Choice.(*ie.GNBIDForGNBID).Value.Bytes
		if bytes.Equal(id, []byte{0, 0, 9}) {
			reached.Do(func() { observed <- fgc })
			return false, gate.Wait(fgc.Context())
		}
		return false, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type result struct {
		fixture *testkit.Fixture
		err     error
	}
	completed := make(chan result, 1)
	go func() {
		fixture, err := testkit.Start(ctx, builder, 2)
		completed <- result{fixture, err}
	}()
	var fgc *core.Aio5gc
	select {
	case fgc = <-observed:
	case <-time.After(5 * time.Second):
		t.Fatal("second native NG Setup never reached the owned gate")
	}
	cancelledAt := time.Now()
	cancel()
	select {
	case outcome := <-completed:
		require.ErrorIs(t, outcome.err, context.Canceled)
		require.Nil(t, outcome.fixture)
	case <-time.After(time.Second):
		t.Fatal("cancelled partial construction did not finish within one second")
	}
	t.Logf("native partial startup cancellation joined in %s", time.Since(cancelledAt))
	require.Zero(t, fgc.ResourceCount(), "core listener and accepted associations must be closed")
	require.Zero(t, fgc.GetAMFContext().GNBCount())
	require.Empty(t, fgc.Snapshot().Errors)
}

func TestNativeAIOAcceptWithClosedStdinOwnsClosableDescriptor(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TEST_ACCEPT_CLOSED_STDIN") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeAIOAcceptWithClosedStdinOwnsClosableDescriptor$", "-test.v", "-test.timeout=5s")
		cmd.Env = append(os.Environ(), "PACKETRUSHER_TEST_ACCEPT_CLOSED_STDIN=1")
		output, err := cmd.CombinedOutput()
		t.Logf("owned AIO accept child with closed stdin:\n%s", output)
		require.NoError(t, err)
		return
	}
	listener, err := service.Listen(netip.MustParseAddrPort("127.0.0.1:0"))
	require.NoError(t, err)
	defer listener.Close()
	address, err := sctp.ResolveSCTPAddr("sctp", listener.Addr().String())
	require.NoError(t, err)
	peer, err := sctp.DialSCTP("sctp", nil, address)
	require.NoError(t, err)
	defer peer.Close()
	require.NoError(t, unix.Close(0))
	accepted, err := listener.Accept()
	require.NoError(t, err)
	defer accepted.Close()
	_, err = unix.FcntlInt(0, unix.F_GETFD, 0)
	require.ErrorIs(t, err, unix.EBADF, "accepted socket must not be transferred as fd zero")
	require.NoError(t, accepted.Close(), "owned core shutdown must close the accepted socket")
}

func TestNativeStartupDeadlineReturnsErrorAndJoinsWaitGroup(t *testing.T) {
	gate := testkit.NewGate()
	fgc, err := new(aio5gc.FiveGCBuilder).WithConfig(testkit.LocalConfig()).WithNGAPDispatcherHook(func(msg message.Message, _ *core.GNBContext, fgc *core.Aio5gc) (bool, error) {
		if _, ok := msg.(*message.NGSetupRequest); ok {
			return true, gate.Wait(fgc.Context())
		}
		return false, nil
	}).Build()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fgc.Close()) })
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	startedAt := time.Now()
	gnbs, err := tools.CreateGnbsContext(ctx, 1, fgc.Config(), &wg)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Empty(t, gnbs)
	// This is a total startup budget, including transport. A lost/delayed
	// first packet need not reach the protocol gate before that budget expires.
	// The separate cancellation test proves teardown after actual gate entry.
	require.Less(t, time.Since(startedAt), time.Second)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("failed startup retained a signal/transport lifecycle slot")
	}
	require.NoError(t, fgc.Close())
	require.Zero(t, fgc.ResourceCount())
	require.Empty(t, fgc.Snapshot().Errors)
}

func TestNativeStartupIPv6CompletesDecodedNGSetup(t *testing.T) {
	cfg := testkit.LocalConfig()
	cfg.GNodeB.ControlIF.AddrPort = netip.MustParseAddrPort("[::1]:0")
	cfg.GNodeB.DataIF.AddrPort = netip.MustParseAddrPort("[::1]:2152")
	cfg.AMFs[0].AddrPort = netip.MustParseAddrPort("[::1]:0")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fixture, err := testkit.Start(ctx, new(aio5gc.FiveGCBuilder).WithConfig(cfg), 1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, fixture.Close()) })
	for _, node := range fixture.Gnbs {
		require.True(t, node.NGSetupReady())
		for amf := range node.IterGnbAmf() {
			require.Equal(t, gnbContext.Active, amf.GetState())
			require.NotZero(t, amf.GetLocalPort())
		}
	}
	require.Empty(t, fixture.Core.Snapshot().Errors)
	require.Len(t, fixture.Core.Snapshot().Associations, 1)
}
