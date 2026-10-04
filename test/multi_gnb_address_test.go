/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package test

import (
	"bytes"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/test/aio5gc"
	coreContext "my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
)

// #138: a gNB whose NG Setup is retried moves to the next address, which the
// gNB created after it must not reuse, whichever AMF the retry happened on.
func TestGnbsKeepDistinctAddressesAfterNGSetupRetry(t *testing.T) {
	for i, tc := range []struct {
		name    string
		retried byte // last octet of the gNB ID whose first NG Setup is dropped
		amfs    int
	}{{"no retry", 0, 1}, {"first gNB retries", 8, 1}, {"second gNB retries", 9, 1}, {"first gNB retries with two AMFs", 8, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			n2 := netip.AddrPortFrom(netip.MustParseAddr("127.0.0.10"), uint16(9498+20*i))
			n3 := netip.AddrPortFrom(n2.Addr(), uint16(2159+i))
			var amfs []*config.AMF
			for a := range tc.amfs {
				amfs = append(amfs, &config.AMF{IPv4Port: config.IPv4Port{AddrPort: netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(38538+2*i+a))}})
			}
			conf := amfTools.GenerateDefaultConf(n2, n3, amfs)
			var dropped atomic.Bool
			core, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).
				WithNGAPDispatcherHook(func(m message.Message, _ *coreContext.GNBContext, _ *coreContext.Aio5gc) (bool, error) {
					setup, ok := m.(*message.NGSetupRequest)
					if !ok {
						return false, nil
					}
					id := setup.GlobalRANNodeID.Choice.(*ie.GlobalGNBID).GNBID.Choice.(*ie.GNBIDForGNBID).Value.Bytes
					return bytes.Equal(id, []byte{0, 0, tc.retried}) && dropped.CompareAndSwap(false, true), nil
				}).Build()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, core.Close()) })

			var wg sync.WaitGroup
			gnbs, err := tools.CreateGnbs(t.Context(), 2, conf, &wg) // GenerateDefaultConf starts at gNB 000008
			require.NoError(t, err)
			require.Len(t, gnbs, 2)
			n2Owner, n3Owner := map[string]string{}, map[netip.Addr]string{}
			for id, g := range gnbs {
				t.Cleanup(g.Terminate)
				require.NotContains(t, n3Owner, g.GetN3GnbIp(), "gNBs share an N3 address")
				n3Owner[g.GetN3GnbIp()] = id
				for amf := range g.IterGnbAmf() {
					if amf.GetState() != gnbcontext.Active {
						continue
					}
					local := amf.GetSCTPConn().LocalAddr().(*sctp.SCTPAddr).IPAddrs[0].IP.String()
					if owner, used := n2Owner[local]; used {
						require.Equal(t, owner, id, "gNBs share N2 address %s", local)
					}
					n2Owner[local] = id
				}
			}
			require.Equal(t, tc.retried != 0, dropped.Load())
		})
	}
}
