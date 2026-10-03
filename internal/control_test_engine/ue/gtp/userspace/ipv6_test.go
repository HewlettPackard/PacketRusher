// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	ipv6 "my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"
	"net/netip"
	"testing"
	"time"
)

func TestIPv6AllocatedIIDPrefixAndRouterAdvertisementIsolation(t *testing.T) {
	peer := udpPeer(t)
	pipe := newPacketPipe()
	registry := NewRegistry()
	cfg := testConfig("127.88.0.1", peer, 7, "10.0.0.1")
	cfg.IPv4 = netip.Addr{}
	cfg.AllowIPv6 = true
	cfg.IPv6InterfaceID = [8]byte{0, 0, 0, 0, 0, 0, 0, 7}
	cfg.IPv6PrefixAllowed = netip.MustParsePrefix("2001:db8:1234::/64").Contains
	session, err := registry.Open(pipe, cfg)
	require.NoError(t, err)
	defer session.Close()
	require.NoError(t, session.Send(ipv6.RouterSolicitation(cfg.IPv6InterfaceID)))
	_, address := packetFrom(t, peer)
	ra, err := hex.DecodeString("6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000")
	require.NoError(t, err)
	wire, err := Encode(cfg.DownlinkTEID, 9, ra)
	require.NoError(t, err)
	_, err = peer.WriteToUDPAddrPort(wire, address)
	require.NoError(t, err)
	select {
	case received := <-session.Advertisements():
		require.Equal(t, ra, received)
	case <-time.After(time.Second):
		t.Fatal("RA not delivered to prefix discovery")
	}
	select {
	case <-pipe.write:
		t.Fatal("RA must not reach host SLAAC")
	case <-time.After(15 * time.Millisecond):
	}
	extended := append([]byte(nil), ra[:40]...)
	extended[6] = 0
	extended[5] += 8
	extended = append(extended, 58, 0, 0, 0, 0, 0, 0, 0)
	extended = append(extended, ra[40:]...)
	wire, err = Encode(cfg.DownlinkTEID, 9, extended)
	require.NoError(t, err)
	_, err = peer.WriteToUDPAddrPort(wire, address)
	require.NoError(t, err)
	select {
	case received := <-session.Advertisements():
		require.Equal(t, extended, received)
	case <-time.After(time.Second):
		t.Fatal("extended RA must be routed to strict prefix validation")
	}
	select {
	case <-pipe.write:
		t.Fatal("extended RA must not bypass prefix discovery through host SLAAC")
	case <-time.After(15 * time.Millisecond):
	}
	packet := make([]byte, 48)
	packet[0] = 0x60
	packet[5] = 8
	packet[6] = 17
	packet[7] = 64
	copy(packet[8:24], netip.MustParseAddr("2001:db8:1234::7").AsSlice())
	copy(packet[24:40], netip.MustParseAddr("2001:db8:ffff::9").AsSlice())
	require.NoError(t, session.Send(packet))
	_, _ = packetFrom(t, peer)
	copy(packet[8:24], netip.MustParseAddr("2001:db8:5678::7").AsSlice())
	require.Error(t, session.Send(packet), "same IID with unallocated prefix must be rejected")
	copy(packet[8:24], netip.MustParseAddr("2001:db8:1234::8").AsSlice())
	require.Error(t, session.Send(packet), "another UE's IID must be rejected")
}
