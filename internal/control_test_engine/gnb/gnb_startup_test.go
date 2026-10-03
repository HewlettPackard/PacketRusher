// SPDX-License-Identifier: Apache-2.0
package gnb

import (
	"context"
	"my5G-RANTester/config"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func waitStartupSlot(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("failed startup retained the caller's lifecycle slot")
	}
}

func TestInitGnbContextFailureConsumesLifecycleSlot(t *testing.T) {
	for _, kind := range []string{"cancelled", "missing-amf", "invalid-identity"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cfg := config.Config{}
			if kind == "cancelled" {
				cancel()
			}
			if kind == "invalid-identity" {
				cfg.AMFs = []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38412")}}}
				cfg.GNodeB.PlmnList.GnbId = "bad identity"
			}
			var wg sync.WaitGroup
			wg.Add(1)
			gnb, err := InitGnbContext(ctx, cfg, &wg)
			require.Error(t, err)
			require.Nil(t, gnb)
			if kind == "cancelled" {
				require.ErrorIs(t, err, context.Canceled)
			}
			waitStartupSlot(t, &wg)
		})
	}
}

func TestLegacyInitGnbInvalidIdentityReturnsWithoutFatalExit(t *testing.T) {
	cfg := config.Config{}
	cfg.AMFs = []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38412")}}}
	cfg.GNodeB.PlmnList.GnbId = "invalid"
	var wg sync.WaitGroup
	wg.Add(1)
	require.Nil(t, InitGnb(cfg, &wg))
	waitStartupSlot(t, &wg)
}
