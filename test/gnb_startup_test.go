/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package test

import (
	stdcontext "context"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/test/aio5gc"
	"my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
	"time"

	ngapType "github.com/free5gc/ngap/message"
	"github.com/stretchr/testify/require"
)

// Nothing listens on the AMF's address: InitGnb gives its caller an error once its
// retries are over, where it used to exit the process.
func TestInitGnbReturnsErrorWhenAmfIsUnreachable(t *testing.T) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:9089"), netip.MustParseAddrPort("127.0.0.1:2152"),
		[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:1")}}})

	start := time.Now()
	_, err := gnb.InitGnb(t.Context(), conf, &sync.WaitGroup{})
	require.ErrorContains(t, err, "127.0.0.1:1")
	require.Less(t, time.Since(start), 5*time.Second)
}

// The AMF accepts the association but does not answer NG Setup. Cancelling the context
// ends the startup at once, without waiting for the NG Setup timeout or the retries, and
// leaves no gNB running.
func TestCreateGnbsStopsWhenCancelledDuringNgSetup(t *testing.T) {
	conf, setupRequests := startAmf(t, "127.0.0.1:9189", false)
	ctx, cancel := stdcontext.WithCancel(t.Context())
	defer cancel()

	result := make(chan error, 1)
	go func() {
		_, err := tools.CreateGnbs(ctx, 1, conf, &sync.WaitGroup{})
		result <- err
	}()
	amfSideGnb := <-setupRequests
	cancel()

	select {
	case err := <-result:
		require.ErrorIs(t, err, stdcontext.Canceled)
	case <-time.After(time.Second):
		t.Fatal("CreateGnbs did not return when its context was cancelled")
	}
	requireAssociationClosed(t, amfSideGnb)
}

// CreateGnbs gives each gNB the next N2 address. Here the second one is not an address
// of this host, so that gNB cannot start: CreateGnbs returns the error and terminates
// the first gNB, which had completed NG Setup.
func TestCreateGnbsTerminatesStartedGnbsOnFailure(t *testing.T) {
	conf, setupRequests := startAmf(t, "127.255.255.254:9289", true)

	gnbs, err := tools.CreateGnbs(t.Context(), 2, conf, &sync.WaitGroup{})
	require.Error(t, err)
	require.Empty(t, gnbs)
	requireAssociationClosed(t, <-setupRequests)
}

// startAmf starts a mock AMF on a free port and returns the configuration of a gNB that
// connects to it from n2. Each NG Setup Request is reported with the AMF's context of the
// gNB that sent it; it is answered only if answerSetup is set.
func startAmf(t *testing.T, n2 string, answerSetup bool) (config.Config, <-chan *context.GNBContext) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort(n2), netip.MustParseAddrPort("127.0.0.1:2152"),
		[]*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:0")}}})

	setupRequests := make(chan *context.GNBContext, 8)
	fiveGC, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).
		WithNGAPDispatcherHook(func(msg ngapType.Message, gnb *context.GNBContext, _ *context.Aio5gc) (bool, error) {
			_, isSetup := msg.(*ngapType.NGSetupRequest)
			if isSetup {
				setupRequests <- gnb
			}
			return isSetup && !answerSetup, nil
		}).
		Build()
	require.NoError(t, err)
	t.Cleanup(func() { _ = fiveGC.Close() })

	conf.AMFs = []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: fiveGC.Addr()}}}
	return conf, setupRequests
}

// requireAssociationClosed checks, on the AMF's side, that the gNB closed its association.
func requireAssociationClosed(t *testing.T, amfSideGnb *context.GNBContext) {
	require.Eventually(t, func() bool { return amfSideGnb.GetSCTPConn().RemoteAddr() == nil }, 2*time.Second, 20*time.Millisecond,
		"the gNB should have closed its association with the AMF")
}
