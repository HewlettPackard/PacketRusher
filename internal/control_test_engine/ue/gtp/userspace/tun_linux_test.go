//go:build linux

// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"
)

func checksum(b []byte) uint16 {
	var sum uint32
	for len(b) > 1 {
		sum += uint32(binary.BigEndian.Uint16(b))
		b = b[2:]
	}
	if len(b) > 0 {
		sum += uint32(b[0]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}
func udpReply(request []byte) []byte {
	ihl := int(request[0]&15) * 4
	data := []byte("pong")
	reply := make([]byte, 28+len(data))
	reply[0] = 0x45
	binary.BigEndian.PutUint16(reply[2:4], uint16(len(reply)))
	reply[8], reply[9] = 64, 17
	copy(reply[12:16], request[16:20])
	copy(reply[16:20], request[12:16])
	binary.BigEndian.PutUint16(reply[10:12], checksum(reply[:20]))
	copy(reply[20:22], request[ihl+2:ihl+4])
	copy(reply[22:24], request[ihl:ihl+2])
	binary.BigEndian.PutUint16(reply[24:26], uint16(8+len(data)))
	copy(reply[28:], data)
	return reply // IPv4 permits a zero UDP checksum.
}
func TestTUNRealUDPTrafficAndCleanup(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("run explicitly in an isolated network namespace with CAP_NET_ADMIN")
	}
	port, link, err := NewTUN("prutun0")
	require.NoError(t, err)
	registry := NewRegistry()
	peer := udpPeer(t)
	cfg := testConfig("127.88.3.1", peer, 44, "10.155.0.2")
	session, err := registry.Open(port, cfg)
	require.NoError(t, err)
	defer session.Close()
	address, _ := netlink.ParseAddr("10.155.0.2/32")
	require.NoError(t, netlink.AddrAdd(link, address))
	require.NoError(t, netlink.LinkSetMTU(link, 1456))
	require.NoError(t, netlink.LinkSetUp(link))
	_, network, _ := net.ParseCIDR("203.0.113.0/24")
	require.NoError(t, netlink.RouteAdd(&netlink.Route{Dst: network, LinkIndex: link.Attrs().Index, Scope: netlink.SCOPE_LINK, Src: net.IPv4(10, 155, 0, 2)}))
	app, err := net.DialUDP("udp4", &net.UDPAddr{IP: net.IPv4(10, 155, 0, 2)}, &net.UDPAddr{IP: net.IPv4(203, 0, 113, 9), Port: 8888})
	require.NoError(t, err)
	defer app.Close()
	require.NoError(t, app.SetDeadline(time.Now().Add(3*time.Second)))
	_, err = app.Write([]byte("ping"))
	require.NoError(t, err)
	wire, source := packetFrom(t, peer)
	id, ip, err := Decode(wire)
	require.NoError(t, err)
	require.Equal(t, cfg.UplinkTEID, id)
	require.Equal(t, cfg.Local, source.Addr())
	require.Equal(t, byte(9), wire[14])
	require.Equal(t, "ping", string(ip[int(ip[0]&15)*4+8:]))
	downlink(t, peer, cfg.Local, cfg.DownlinkTEID, udpReply(ip))
	result := make([]byte, 16)
	n, err := app.Read(result)
	require.NoError(t, err)
	require.Equal(t, "pong", string(result[:n]))
	// The endpoint identity remains the same when the target N3/TEIDs change.
	target := cfg
	target.Local = netip.MustParseAddr("127.88.3.2")
	target.UplinkTEID = 145
	target.DownlinkTEID = 45
	require.NoError(t, session.Update(target))
	_, err = app.Write([]byte("ping"))
	require.NoError(t, err)
	wire, source = packetFrom(t, peer)
	id, ip, err = Decode(wire)
	require.NoError(t, err)
	require.Equal(t, target.UplinkTEID, id)
	require.Equal(t, target.Local, source.Addr())
	downlink(t, peer, target.Local, target.DownlinkTEID, udpReply(ip))
	n, err = app.Read(result)
	require.NoError(t, err)
	require.Equal(t, "pong", string(result[:n]))
	session.Close()
	_, err = netlink.LinkByName("prutun0")
	require.Error(t, err)
	rebound, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(target.Local, 2152)))
	require.NoError(t, err)
	rebound.Close()
}
