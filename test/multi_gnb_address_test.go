// SPDX-License-Identifier: Apache-2.0
package test

import (
	"bytes"
	"encoding/binary"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	mobility "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_mobility_management"
	codec "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc"
	coreContext "my5G-RANTester/test/aio5gc/context"
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
)

// #138 reported two NG Setups sharing an IP and a stale Path Switch N3 address.
// Exercise actual production gNB creation and SCTP binds, then decode the real
// target Path Switch transfer. Loopback /8 requires no network alias changes.
func TestMultiGNBUsesDistinctN2N3AndTargetPathSwitchAddress(t *testing.T) {
	t.Run("normal", func(t *testing.T) { multiGNBAddressProof(t, false) })
	t.Run("after-ng-setup-retry", func(t *testing.T) { multiGNBAddressProof(t, true) })
}

func multiGNBAddressProof(t *testing.T, retry bool) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.10:9498"), netip.MustParseAddrPort("127.0.0.10:2159"), []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38538")}},
	})
	var dropped atomic.Bool
	builder := (&aio5gc.FiveGCBuilder{}).WithConfig(conf)
	if retry {
		builder.WithNGAPDispatcherHook(func(m message.Message, _ *coreContext.GNBContext, _ *coreContext.Aio5gc) (bool, error) {
			if setup, ok := m.(*message.NGSetupRequest); ok {
				node := setup.GlobalRANNodeID.Choice.(*ie.GlobalGNBID)
				id := node.GNBID.Choice.(*ie.GNBIDForGNBID).Value.Bytes
				if bytes.Equal(id, []byte{0, 0, 9}) && dropped.CompareAndSwap(false, true) {
					return true, nil
				}
			}
			return false, nil
		})
	}
	core, err := builder.Build()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, core.Close()) })
	var wg sync.WaitGroup
	gnbs := tools.CreateGnbs(2, conf, &wg)
	t.Cleanup(func() {
		for _, g := range gnbs {
			g.Terminate()
		}
	})
	require.Len(t, gnbs, 2)
	first, target := gnbs["000008"], gnbs["000009"]
	// GenerateDefaultConf uses 000008; fail explicitly if the fixture changes.
	require.NotNil(t, first)
	require.NotNil(t, target)
	require.NotEqual(t, first.GetGnbIpPort().Addr(), target.GetGnbIpPort().Addr(), "N2 associations must use distinct addresses")
	require.NotEqual(t, first.GetN3GnbIp(), target.GetN3GnbIp(), "N3 endpoints must use distinct addresses")
	if retry {
		require.True(t, dropped.Load(), "the target's first NG Setup must have been withheld")
		require.NotEqual(t, conf.GNodeB.DataIF.WithNextAddr().Addr(), target.GetN3GnbIp(), "the retry must move the target address")
	}
	for _, g := range []*gnbcontext.GNBContext{first, target} {
		want := g.GetGnbIpPort().Addr()
		require.True(t, want.IsLoopback())
		require.Equal(t, want, g.GetN3GnbIp())
		require.Equal(t, uint16(9498), g.GetGnbIpPort().Port())
		for a := range g.IterGnbAmf() {
			require.Equal(t, gnbcontext.Active, a.GetState())
			local := a.GetSCTPConn().LocalAddr().(*sctp.SCTPAddr)
			require.Equal(t, want.String(), local.IPAddrs[0].IP.String())
			peer, err := core.GetAMFContext().GetGnb(local.String())
			require.NoError(t, err)
			require.NotNil(t, peer.GetGlobalRanNodeID().GNbId)
			require.Equal(t, g.GetGnbId(), peer.GetGlobalRanNodeID().GNbId.GNBValue)
		}
	}
	ue, err := target.NewGnBUe(make(chan gnbcontext.UEMessage, 2), make(chan gnbcontext.UEMessage, 2), 42, nil)
	require.NoError(t, err)
	ue.SetAmfUeId(42)
	algorithms := aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"000001"}, &ie.UESecurityCapabilities{
		NRencryptionAlgorithms:             &ie.NRencryptionAlgorithms{Value: algorithms},
		NRintegrityProtectionAlgorithms:    &ie.NRintegrityProtectionAlgorithms{Value: algorithms},
		EUTRAencryptionAlgorithms:          &ie.EUTRAencryptionAlgorithms{Value: algorithms},
		EUTRAintegrityProtectionAlgorithms: &ie.EUTRAintegrityProtectionAlgorithms{Value: algorithms},
	})
	_, err = ue.CreatePduSession(1, "127.0.0.20", "01", "000001", 0, 9, 1, 9, 100, 200)
	require.NoError(t, err)
	wire, err := mobility.PathSwitchRequest(target, ue)
	require.NoError(t, err)
	decoded, err := message.Parse(wire)
	require.NoError(t, err)
	request := decoded.(*message.PathSwitchRequest)
	require.Len(t, request.PDUSessionResourceToBeSwitchedDLList.List, 1)
	transfer := &ie.PathSwitchRequestTransfer{}
	require.NoError(t, codec.Unmarshal(request.PDUSessionResourceToBeSwitchedDLList.List[0].PathSwitchRequestTransfer, transfer))
	tunnel, err := codec.Tunnel(transfer.DLNGUUPTNLInformation)
	require.NoError(t, err)
	require.Equal(t, target.GetN3GnbIp().AsSlice(), tunnel.TransportLayerAddress.Value.Bytes, "Path Switch must encode the current target, including after retries")
	require.Equal(t, uint32(200), binary.BigEndian.Uint32(tunnel.GTPTEID.Value))
}
