//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"testing"
)

func TestIPv6AllocationCommitFailureDeactivatesAndRetiresEndpoint(t *testing.T) {
	r, state, c := isolatedRegistry()
	c.AllowIPv6 = true
	c.IPv6InterfaceID[7] = 7
	s, err := r.Open(c)
	require.NoError(t, err)
	first := netip.MustParseAddr("2001:db8:1::7")
	require.NoError(t, s.SetIPv6(first))
	require.Equal(t, uint8(1), state.values["sessions"][c.identity()].(binding).PrefixValid)
	require.Error(t, s.SetIPv6(netip.MustParseAddr("2001:db8:1::8")), "unallocated IID must fail before changing the canonical map")
	require.Equal(t, first, s.cfg.IPv6)
	injected := errors.New("map replacement failed")
	state.fail = func(op, name string, k, v any) error {
		if op == "put" && name == "sessions" {
			return injected
		}
		return nil
	}
	require.ErrorIs(t, s.SetIPv6(netip.MustParseAddr("2001:db8:2::7")), injected)
	require.NotContains(t, state.values["sessions"], c.identity(), "old authorization cannot survive callback failure")
	require.Empty(t, *r.controls.Load())
	require.Error(t, s.SetIPv6(first), "a failed allocation owner cannot re-enable itself")
	state.fail = nil
	require.NoError(t, s.SetIPv6(netip.Addr{}))
	require.NoError(t, s.Close())
	require.NoError(t, s.SetIPv6(netip.Addr{}), "revocation stays idempotent after final cleanup")
	require.Empty(t, r.sessions)
	require.Empty(t, r.locals)
	require.Empty(t, r.links)
}
func TestIPv6RevocationFailureRetainsCleanupClaim(t *testing.T) {
	r, state, c := isolatedRegistry()
	c.AllowIPv6 = true
	c.IPv6InterfaceID[7] = 7
	s, err := r.Open(c)
	require.NoError(t, err)
	require.NoError(t, s.SetIPv6(netip.MustParseAddr("2001:db8:1::7")))
	fail := errors.New("map unavailable")
	state.fail = func(op, name string, k, v any) error {
		if name == "sessions" {
			return fail
		}
		return nil
	}
	require.ErrorIs(t, s.SetIPv6(netip.Addr{}), fail)
	require.True(t, s.stopping)
	require.NotEmpty(t, s.leases)
	require.NotNil(t, s.endpoint)
	require.Empty(t, *r.controls.Load())
	require.ErrorIs(t, s.Close(), fail)
	state.fail = nil
	require.NoError(t, s.SetIPv6(netip.Addr{}))
	require.NoError(t, s.Close())
	require.Empty(t, r.sessions)
	require.Empty(t, r.locals)
}
func TestHandoverPreservesCurrentPrefixAndUsesValidatedNextHop(t *testing.T) {
	r, state, c := isolatedRegistry()
	c.AllowIPv6 = true
	c.IPv6InterfaceID[7] = 7
	gateway := netip.MustParseAddr("10.88.0.254")
	r.discover = func(Config) (n3Path, error) { return n3Path{IfIndex: 10, NextHop: gateway}, nil }
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	prefix := netip.MustParseAddr("2001:db8:2::7")
	require.NoError(t, s.SetIPv6(prefix))
	next := c
	next.Local = netip.MustParseAddr("10.88.0.3")
	next.DownlinkTEID++
	next.UplinkTEID++
	require.NoError(t, s.Update(next))
	canonical := state.values["sessions"][c.identity()].(binding)
	require.Equal(t, ipv4(gateway), canonical.NextHop)
	require.Equal(t, prefix, s.cfg.IPv6, "stale handover snapshot cannot revoke callback-owned prefix")
	require.Equal(t, s.cfg.binding(s.path), canonical)
	require.NoError(t, s.SetIPv6(netip.Addr{}))
	require.Equal(t, uint8(0), state.values["sessions"][c.identity()].(binding).PrefixValid)
}

func TestRouterAdvertisementAdmissionUsesExactOwnerAndDirection(t *testing.T) {
	r, _, c := isolatedRegistry()
	c.AllowIPv6 = true
	c.IPv6InterfaceID[7] = 7
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	ra, err := hex.DecodeString("6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000")
	require.NoError(t, err)
	wire, err := userspace.Encode(c.DownlinkTEID, c.QFI, ra)
	require.NoError(t, err)
	wire[13] = 0
	peer := netip.AddrPortFrom(c.Remote, 2152)
	r.receiveControl(c.Local, peer, wire)
	select {
	case got := <-s.advertisements:
		require.Equal(t, ra, got)
	default:
		t.Fatal("owned valid advertisement was not queued")
	}
	for _, mutate := range []func([]byte){func(p []byte) { p[13] = 0x10 }, func(p []byte) { p[14]++ }, func(p []byte) { p[7]++ }, func(p []byte) { p[23] = 64 }} {
		changed := append([]byte(nil), wire...)
		mutate(changed)
		r.receiveControl(c.Local, peer, changed)
		require.Empty(t, s.advertisements)
	}
	r.receiveControl(c.Local, netip.MustParseAddrPort("10.88.0.99:2152"), wire)
	require.Empty(t, s.advertisements)
}

type fallbackSender struct {
	fakeCloser
	packets [][]byte
	peers   []netip.AddrPort
}

func TestOptionalDownlinkDispatchChecksCurrentOwnerFlowAndAllocation(t *testing.T) {
	r, _, c := isolatedRegistry()
	var injected [][]byte
	c.Inject = func(p []byte) error { injected = append(injected, append([]byte(nil), p...)); return nil }
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	inner := wirePacket(c)[58:]
	packet := append([]byte{0x37, 255, 0, 0, 0, 0, 0, 0, 0x12, 0x34, 7, 0x40, 1, 8, 0x68, 0x85, 1, 0, c.QFI, 0}, inner...)
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)-8))
	binary.BigEndian.PutUint32(packet[4:8], c.DownlinkTEID)
	peer := netip.AddrPortFrom(c.Remote, 2152)
	r.receiveControl(c.Local, peer, packet)
	require.Equal(t, [][]byte{inner}, injected)
	injected = nil
	for _, mutate := range []func([]byte){
		func(p []byte) { p[18]++ },        // wrong PSC flow hidden behind unknown extension
		func(p []byte) { p[17] = 0x10 },   // uplink PSC direction
		func(p []byte) { p[7]++ },         // old/wrong TEID
		func(p []byte) { p[12] = 0 },      // extension size zero
		func(p []byte) { p[len(p)-13]++ }, // unallocated destination
	} {
		bad := append([]byte(nil), packet...)
		mutate(bad)
		r.receiveControl(c.Local, peer, bad)
		require.Empty(t, injected)
	}
	r.receiveControl(c.Local, netip.MustParseAddrPort("10.88.0.99:2152"), packet)
	r.receiveControl(c.Local, netip.AddrPortFrom(c.Remote, 2153), packet)
	r.receiveControl(netip.MustParseAddr("10.88.0.99"), peer, packet)
	require.Empty(t, injected)
	next := c
	next.Remote = netip.MustParseAddr("10.88.0.4")
	next.DownlinkTEID++
	require.NoError(t, s.Update(next))
	r.receiveControl(c.Local, peer, packet)
	require.Empty(t, injected, "old peer/TEID snapshot cannot dispatch after handover")
	binary.BigEndian.PutUint32(packet[4:8], next.DownlinkTEID)
	r.receiveControl(next.Local, netip.AddrPortFrom(next.Remote, 2152), packet)
	require.Equal(t, [][]byte{inner}, injected)
	require.NoError(t, s.Close())
	injected = nil
	r.receiveControl(next.Local, netip.AddrPortFrom(next.Remote, 2152), packet)
	require.Empty(t, injected, "retired owner cannot accept delegated packets")
}

func (s *fallbackSender) Send(packet []byte, peer netip.AddrPort) error {
	s.packets = append(s.packets, append([]byte(nil), packet...))
	s.peers = append(s.peers, peer)
	return nil
}
func TestJumboUplinkDispatchChecksCurrentMappingAndSource(t *testing.T) {
	r, _, c := isolatedRegistry()
	c.MTU = 65491
	sender := &fallbackSender{}
	r.listen = func(netip.Addr) (closer, error) { return sender, nil }
	s, err := r.Open(c)
	require.NoError(t, err)
	packet := make([]byte, 32000)
	packet[0] = 0x45
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))
	copy(packet[12:16], c.IPv4.AsSlice())
	copy(packet[16:20], []byte{192, 0, 2, 1})
	require.NoError(t, s.SendUplink(packet))
	teid, payload, err := userspace.Decode(sender.packets[0])
	require.NoError(t, err)
	require.Equal(t, c.UplinkTEID, teid)
	require.Equal(t, packet, payload)
	next := c
	next.Remote = netip.MustParseAddr("10.88.0.4")
	next.UplinkTEID++
	require.NoError(t, s.Update(next))
	require.NoError(t, s.SendUplink(packet))
	teid, _, err = userspace.Decode(sender.packets[1])
	require.NoError(t, err)
	require.Equal(t, next.UplinkTEID, teid)
	require.Equal(t, netip.AddrPortFrom(next.Remote, 2152), sender.peers[1])
	packet[12]++
	require.Error(t, s.SendUplink(packet))
	require.Len(t, sender.packets, 2)
	require.Error(t, s.SendUplink(make([]byte, 65535)))
	require.Error(t, s.SendUplink(make([]byte, 8000)))
	require.NoError(t, s.Close())
	require.ErrorIs(t, s.SendUplink(packet), net.ErrClosed)
}
