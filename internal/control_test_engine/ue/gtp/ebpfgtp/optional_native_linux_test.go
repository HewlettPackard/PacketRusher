//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

// The same actual UDP application, TUN, complete outer packets and deadline
// prove supported optional formats and session admission for both backends.
func TestNativeOptionalHeadersBothBackendsAndRecoveryEcho(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	for _, backend := range []string{"userspace", "ebpf"} {
		t.Run(backend, func(t *testing.T) {
			lo, err := netlink.LinkByName("lo")
			require.NoError(t, err)
			for _, cidr := range []string{"127.88.5.1/32", "127.88.5.9/32"} {
				a, err := netlink.ParseAddr(cidr)
				require.NoError(t, err)
				require.NoError(t, netlink.AddrAdd(lo, a))
				t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
			}
			port, device, err := userspace.NewTUN("pr-optional")
			require.NoError(t, err)
			defer port.Close()
			require.NoError(t, netlink.LinkSetMTU(device, 1456))
			require.NoError(t, netlink.AddrAdd(device, &netlink.Addr{IPNet: testpeer.Network("10.60.0.1/32")}))
			require.NoError(t, netlink.LinkSetUp(device))
			require.NoError(t, netlink.RouteAdd(&netlink.Route{Dst: testpeer.Network("192.0.2.1/32"), LinkIndex: device.Attrs().Index, Scope: netlink.SCOPE_LINK}))
			c := Config{Local: netip.MustParseAddr("127.88.5.1"), Remote: netip.MustParseAddr("127.88.5.9"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 60, DownlinkTEID: 61, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 1456, Inject: func(p []byte) error { _, err := port.Write(p); return err }}
			var closeSession func()
			if backend == "ebpf" {
				s, err := NewRegistry().Open(c)
				require.NoError(t, err)
				closeSession = func() { require.NoError(t, s.Close()) }
			} else {
				s, err := userspace.NewRegistry().Open(port, userspace.Config{Local: c.Local, Remote: netip.AddrPortFrom(c.Remote, 2152), IPv4: c.IPv4, UplinkTEID: c.UplinkTEID, DownlinkTEID: c.DownlinkTEID, QFI: c.QFI})
				require.NoError(t, err)
				closeSession = s.Close
			}
			defer closeSession()
			app, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IP(c.IPv4.AsSlice()), Port: 50000})
			require.NoError(t, err)
			defer app.Close()
			raw, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.IPPROTO_RAW)
			require.NoError(t, err)
			defer unix.Close(raw)
			inner := wirePacket(c)[58:]
			packet := func(flags byte, optional []byte) []byte {
				p := append([]byte{flags, 255, 0, 0, 0, 0, 0, 61}, optional...)
				p = append(p, inner...)
				binary.BigEndian.PutUint16(p[2:4], uint16(len(p)-8))
				return p
			}
			send := func(gtp []byte, peer netip.Addr) {
				t.Helper()
				outer := append([]byte(nil), wirePacket(c)[14:42]...)
				copy(outer[12:16], peer.AsSlice())
				binary.BigEndian.PutUint16(outer[2:4], uint16(len(outer)+len(gtp)))
				binary.BigEndian.PutUint16(outer[24:26], uint16(8+len(gtp)))
				outer[10], outer[11] = 0, 0
				binary.BigEndian.PutUint16(outer[10:12], testpeer.Checksum(outer[:20]))
				require.NoError(t, unix.Sendto(raw, append(outer, gtp...), 0, &unix.SockaddrInet4{Addr: c.Local.As4()}))
			}
			for _, p := range [][]byte{
				packet(0x31, []byte{0, 0, 7, 0}),
				packet(0x37, []byte{0x12, 0x34, 7, 0x40, 1, 8, 0x68, 0x85, 1, 0, 9, 0}),
				packet(0x34, []byte{0, 0, 0, 0x40, 1, 8, 0x68, 0}),
				packet(0x34, []byte{0, 0, 0, 0x85, 1, 0, 0x49, 0}),
				packet(0x34, []byte{0, 0, 0, 0x85, 5, 0x0f, 0xc9, 0xe0, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 1, 2, 3, 4, 0}), // all optional fields
			} {
				send(p, c.Remote)
				require.NoError(t, app.SetReadDeadline(time.Now().Add(time.Second)))
				b := make([]byte, 32)
				n, _, err := app.ReadFromUDP(b)
				require.NoError(t, err)
				require.Equal(t, "wire", string(b[:n]))
			}
			for _, optional := range [][]byte{
				{0, 0, 0, 0x40, 1, 8, 0x68, 0x85, 1, 0, 8, 0},    // hidden wrong QFI
				{0, 0, 0, 0x40, 1, 8, 0x68, 0x85, 1, 0x10, 9, 0}, // hidden UL direction
				{0, 0, 0, 0x85, 1, 8, 9, 0},                      // truncated optional timestamp
				{0, 0, 0, 0x85, 1, 0, 9, 0x85, 1, 0, 9, 0},       // duplicate PSC
				{0, 0, 0, 0x40, 0, 0, 0, 0},                      // zero extension size
				{0, 0, 0, 0x40, 255, 0, 0, 0},                    // extension past datagram
			} {
				send(packet(0x34, optional), c.Remote)
			}
			valid := packet(0x37, []byte{0, 0, 7, 0x85, 1, 0, 9, 0})
			wrong := append([]byte(nil), valid...)
			wrong[7]++
			send(wrong, c.Remote)
			send(valid, netip.MustParseAddr("127.88.5.8"))
			wrong = append([]byte(nil), valid...)
			wrong[len(wrong)-13]++
			start := len(wrong) - len(inner)
			wrong[start+10], wrong[start+11] = 0, 0
			binary.BigEndian.PutUint16(wrong[start+10:start+12], testpeer.Checksum(wrong[start:start+20]))
			send(wrong, c.Remote) // wrong inner destination
			// A final positive is an ordered barrier for this socket's rejected
			// datagrams; it must not be displaced by an unauthorized delivery.
			marker := append([]byte(nil), valid...)
			copy(marker[len(marker)-4:], "done")
			send(marker, c.Remote)
			require.NoError(t, app.SetReadDeadline(time.Now().Add(time.Second)))
			b := make([]byte, 32)
			n, _, err := app.ReadFromUDP(b)
			require.NoError(t, err)
			require.Equal(t, "done", string(b[:n]))
			// Echo must traverse actual TC ingress and the management helper.
			peer, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(c.Remote, 2152)))
			require.NoError(t, err)
			defer peer.Close()
			echo := []byte{0x32, 1, 0, 6, 0, 0, 0, 0, 0x12, 0x34, 0, 0, 14, 7}
			bad := append([]byte(nil), echo...)
			bad[3]++
			send(bad, c.Remote)
			{
				other, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.MustParseAddrPort("127.88.5.8:2152")))
				require.NoError(t, err)
				defer other.Close()
				send(echo, netip.MustParseAddr("127.88.5.8"))
				require.NoError(t, other.SetReadDeadline(time.Now().Add(20*time.Millisecond)))
				_, _, err = other.ReadFromUDP(b)
				require.True(t, os.IsTimeout(err), "uncommitted Echo peer must not receive a response")
			}
			send(echo, c.Remote)
			require.NoError(t, peer.SetReadDeadline(time.Now().Add(time.Second)))
			n, _, err = peer.ReadFromUDP(b)
			require.NoError(t, err)
			require.Equal(t, echoResponse(echo), b[:n])
			closeSession() // join every owned receiver before the final emptiness assertion
			require.NoError(t, app.SetReadDeadline(time.Now().Add(20*time.Millisecond)))
			_, _, err = app.ReadFromUDP(b)
			require.True(t, os.IsTimeout(err))
		})
	}
}
