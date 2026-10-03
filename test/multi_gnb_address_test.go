// SPDX-License-Identifier: Apache-2.0
package test

import (
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
	amfTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"testing"
)

// #138 reported two NG Setups sharing an IP and a stale Path Switch N3 address.
// Exercise actual production gNB creation and SCTP binds, then decode the real
// target Path Switch transfer. Loopback /8 requires no network alias changes.
func TestMultiGNBUsesDistinctN2N3AndTargetPathSwitchAddress(t *testing.T) {
	conf := amfTools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.10:9498"), netip.MustParseAddrPort("127.0.0.10:2159"), []*config.AMF{
		{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:38538")}},
	})
	core, err := (&aio5gc.FiveGCBuilder{}).WithConfig(conf).Build()
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
	for i, g := range []*gnbcontext.GNBContext{first, target} {
		want := netip.AddrFrom4([4]byte{127, 0, 0, byte(10 + i)})
		require.Equal(t, want, g.GetGnbIpPort().Addr())
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
	require.Equal(t, []byte{127, 0, 0, 11}, tunnel.TransportLayerAddress.Value.Bytes)
	require.Equal(t, uint32(200), binary.BigEndian.Uint32(tunnel.GTPTEID.Value))
}
