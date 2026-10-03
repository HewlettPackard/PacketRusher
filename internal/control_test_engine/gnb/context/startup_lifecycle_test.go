// SPDX-License-Identifier: Apache-2.0
package context

import (
	stdcontext "context"
	"net/netip"
	"testing"
	"time"

	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
)

func TestStartupEndpointRetryPreservesAdmissionAndShutdownOwnership(t *testing.T) {
	gnb := new(GNBContext)
	gnb.NewRanGnbContext("000008", "001", "01", "000001", "1", "", netip.MustParseAddrPort("127.0.0.1:0"), netip.MustParseAddrPort("127.0.0.1:2152"))
	done, inbound := gnb.Done(), gnb.GetInboundChannel()
	workerStarted := make(chan struct{})
	require.True(t, gnb.RunAssociation(func() { close(workerStarted); <-done }))
	<-workerStarted
	gnb.SetStartupEndpoints(netip.MustParseAddrPort("127.0.0.2:0"), netip.MustParseAddrPort("127.0.0.2:2152"))
	require.Equal(t, done, gnb.Done())
	require.Equal(t, inbound, gnb.GetInboundChannel())
	gnb.Terminate()
	gnb.WaitAssociations()
	require.False(t, gnb.RunAssociation(func() { t.Error("admitted after shutdown") }))
	select {
	case <-done:
	default:
		t.Fatal("retry detached the shutdown signal from an admitted worker")
	}
}

func TestWaitActiveObservesResponseOrCancellation(t *testing.T) {
	for _, outcome := range []string{"response", "cancel", "stop"} {
		t.Run(outcome, func(t *testing.T) {
			amf := &GNBAmf{}
			ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), time.Second)
			defer cancel()
			stopped := make(chan struct{})
			result := make(chan error, 1)
			go func() { result <- amf.WaitActive(ctx, stopped) }()
			switch outcome {
			case "response":
				amf.SetStateActive()
			case "cancel":
				cancel()
			case "stop":
				close(stopped)
			}
			select {
			case err := <-result:
				if outcome == "response" {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, stdcontext.Canceled)
				}
			case <-time.After(time.Second):
				t.Fatal("NG Setup readiness wait did not observe its completion signal")
			}
		})
	}
}

func TestFailedAssociationCannotClearItsReplacement(t *testing.T) {
	amf := &GNBAmf{}
	old, replacement := &sctp.SCTPConn{}, &sctp.SCTPConn{}
	amf.SetSCTPConn(replacement)
	amf.ClearSCTPConn(old)
	require.Same(t, replacement, amf.GetSCTPConn())
	amf.ClearSCTPConn(replacement)
	require.Nil(t, amf.GetSCTPConn())
}
