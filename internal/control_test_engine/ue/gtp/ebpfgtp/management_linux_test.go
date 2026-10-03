//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNativeManagementEchoAndJoinedClose(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private namespace")
	}
	local := netip.MustParseAddr("127.0.0.1")
	peer := netip.MustParseAddr("127.0.0.2")
	management, err := newManagementSocket(local, func(p netip.AddrPort) bool { return p == netip.AddrPortFrom(peer, 2152) })
	require.NoError(t, err)
	defer management.Close()
	socket, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(peer, 2152)))
	require.NoError(t, err)
	defer socket.Close()
	require.NoError(t, socket.SetDeadline(time.Now().Add(2*time.Second)))
	request := []byte{0x32, 1, 0, 4, 0, 0, 0, 0, 0x12, 0x34, 0, 0}
	_, err = socket.WriteToUDPAddrPort(request, netip.AddrPortFrom(local, 2152))
	require.NoError(t, err)
	packet := make([]byte, 100)
	n, source, err := socket.ReadFromUDPAddrPort(packet)
	require.NoError(t, err)
	require.Equal(t, netip.AddrPortFrom(local, 2152), source)
	require.Equal(t, echoResponse(request), packet[:n])
	require.NoError(t, management.Close())
	select {
	case <-management.done:
	default:
		t.Fatal("management worker was not joined")
	}
	rebound, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(local, 2152)))
	require.NoError(t, err)
	rebound.Close()
}
func TestCommittedPeerSnapshotTracksHandoverAndRelease(t *testing.T) {
	r, _, c := isolatedRegistry()
	s, err := r.Open(c)
	require.NoError(t, err)
	require.True(t, (*r.peers.Load())[peerKey{c.Local, c.Remote}])
	next := c
	next.Remote = netip.MustParseAddr("10.88.0.4")
	next.DownlinkTEID++
	require.NoError(t, s.Update(next))
	require.False(t, (*r.peers.Load())[peerKey{c.Local, c.Remote}])
	require.True(t, (*r.peers.Load())[peerKey{next.Local, next.Remote}])
	require.NoError(t, s.Close())
	require.Empty(t, *r.peers.Load())
}
