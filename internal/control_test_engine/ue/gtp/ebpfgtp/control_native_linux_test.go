//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"
)

func TestNativeLoopbackRouterAdvertisementAdmission(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		address, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, address))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, address)) })
	}
	port, device, err := userspace.NewTUN("pr-v6-control")
	require.NoError(t, err)
	defer port.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 1456))
	require.NoError(t, netlink.LinkSetUp(device))
	r := NewRegistry()
	cfg := Config{Local: netip.MustParseAddr("127.88.4.1"), Remote: netip.MustParseAddr("127.88.4.9"), AllowIPv6: true, IPv6InterfaceID: [8]byte{0, 0, 0, 0, 0, 0, 0, 7}, UplinkTEID: 60, DownlinkTEID: 61, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 1456}
	peer, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(cfg.Remote, 2152)))
	require.NoError(t, err)
	defer peer.Close()
	raw, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, int(0x0300))
	require.NoError(t, err)
	defer unix.Close(raw)
	require.NoError(t, unix.Bind(raw, &unix.SockaddrLinklayer{Protocol: 0x0300, Ifindex: 1}))
	s, err := r.Open(cfg)
	require.NoError(t, err)
	defer s.Close()
	require.NoError(t, s.SendControl(ipv6.RouterSolicitation(cfg.IPv6InterfaceID)))
	buf := make([]byte, 65535)
	require.NoError(t, peer.SetDeadline(time.Now().Add(time.Second)))
	n, source, err := peer.ReadFromUDPAddrPort(buf)
	require.NoError(t, err)
	_, rs, err := userspace.Decode(buf[:n])
	require.NoError(t, err)
	require.Equal(t, byte(133), rs[40])
	ra, _ := hex.DecodeString("6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000")
	wire, err := userspace.Encode(61, 9, ra)
	require.NoError(t, err)
	wire[13] = 0
	_, err = peer.WriteToUDPAddrPort(wire, source)
	require.NoError(t, err)
	// Observe the actual loopback frame before deciding whether a transport
	// failure is admission, checksum offload metadata, or user-space dispatch.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		ready, err := unix.Poll([]unix.PollFd{{Fd: int32(raw), Events: unix.POLLIN}}, 10)
		require.NoError(t, err)
		if ready == 0 {
			continue
		}
		n, _, err := unix.Recvfrom(raw, buf, 0)
		if err == unix.EAGAIN {
			continue
		}
		require.NoError(t, err)
		if n < 58 || binary.BigEndian.Uint32(buf[46:50]) != 61 {
			continue
		}
		pseudo := append([]byte(nil), buf[26:34]...)
		pseudo = append(pseudo, 0, 17, buf[38], buf[39])
		pseudo = append(pseudo, buf[34:n]...)
		t.Logf("captured loopback RA outer UDP checksum=%04x complete residual=%04x len%d", binary.BigEndian.Uint16(buf[40:42]), testpeer.Checksum(pseudo), n)
		break
	}
	select {
	case got := <-s.Advertisements():
		require.Equal(t, ra, got)
	case <-time.After(time.Second):
		ul, dl, drops, err := r.Stats()
		t.Fatalf("actual RA not admitted: kernel UL%d DL%d drops%d err%v", ul, dl, drops, err)
	}
}

func TestNativeKernelChecksumAndReassemblyBeforeFallback(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		a, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, a))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
	}
	port, device, err := userspace.NewTUN("pr-reassembly")
	require.NoError(t, err)
	defer port.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 1456))
	require.NoError(t, netlink.LinkSetUp(device))
	injected := make(chan []byte, 8)
	c := Config{Local: netip.MustParseAddr("127.88.4.1"), Remote: netip.MustParseAddr("127.88.4.9"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 60, DownlinkTEID: 61, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 1456, Inject: func(p []byte) error { injected <- append([]byte(nil), p...); return nil }}
	r := NewRegistry()
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	inner := wirePacket(c)[58:]
	gtp, err := userspace.Encode(61, 9, inner)
	require.NoError(t, err)
	gtp[13] = 0
	udp := make([]byte, 8+len(gtp))
	binary.BigEndian.PutUint16(udp[0:2], 2152)
	binary.BigEndian.PutUint16(udp[2:4], 2152)
	binary.BigEndian.PutUint16(udp[4:6], uint16(len(udp)))
	copy(udp[8:], gtp)
	pseudo := append(c.Remote.AsSlice(), c.Local.AsSlice()...)
	pseudo = append(pseudo, 0, 17, byte(len(udp)>>8), byte(len(udp)))
	pseudo = append(pseudo, udp...)
	sum := testpeer.Checksum(pseudo)
	if sum == 0 {
		sum = 0xffff
	}
	binary.BigEndian.PutUint16(udp[6:8], sum)
	raw, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.IPPROTO_RAW)
	require.NoError(t, err)
	defer unix.Close(raw)
	send := func(id uint16, offset uint16, data []byte) {
		t.Helper()
		ip := make([]byte, 20)
		ip[0] = 0x45
		ip[8] = 64
		ip[9] = 17
		binary.BigEndian.PutUint16(ip[2:4], uint16(20+len(data)))
		binary.BigEndian.PutUint16(ip[4:6], id)
		binary.BigEndian.PutUint16(ip[6:8], offset)
		copy(ip[12:16], c.Remote.AsSlice())
		copy(ip[16:20], c.Local.AsSlice())
		binary.BigEndian.PutUint16(ip[10:12], testpeer.Checksum(ip))
		require.NoError(t, unix.Sendto(raw, append(ip, data...), 0, &unix.SockaddrInet4{Addr: c.Local.As4()}))
	}
	corrupt := append([]byte(nil), udp...)
	corrupt[7] ^= 1
	send(221, 0, corrupt)
	// A correct completed checksum fragmented on outer IPv4 cannot use direct
	// TC decapsulation. The actual UDP stack must reassemble and authorize it.
	send(222, 0x2000, udp[:32])
	send(222, 4, udp[32:])
	select {
	case got := <-injected:
		require.Equal(t, inner, got)
	case <-time.After(time.Second):
		t.Fatal("kernel-validated reassembled owned packet was not injected")
	}
	require.NoError(t, s.Close())
	require.Empty(t, injected, "corrupted completed UDP checksum must never reach the injector")
	t.Log("actual corrupt CHECKSUM_NONE dropped; valid outer fragments reassembled and admitted once; worker joined")
}

func TestNativeJumboChecksumTailsAndQFI(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		a, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, a))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
	}
	port, device, err := userspace.NewTUN("pr-jumbo")
	require.NoError(t, err)
	defer port.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 65491))
	require.NoError(t, netlink.AddrAdd(device, &netlink.Addr{IPNet: testpeer.Network("10.60.0.1/32")}))
	require.NoError(t, netlink.LinkSetUp(device))
	require.NoError(t, netlink.RouteAdd(&netlink.Route{Dst: testpeer.Network("192.0.2.1/32"), LinkIndex: device.Attrs().Index, Scope: netlink.SCOPE_LINK}))
	c := Config{Local: netip.MustParseAddr("127.88.4.1"), Remote: netip.MustParseAddr("127.88.4.9"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 60, DownlinkTEID: 61, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 65491}
	r := NewRegistry()
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	app, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IP(c.IPv4.AsSlice()), Port: 50000})
	require.NoError(t, err)
	defer app.Close()
	raw, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.IPPROTO_RAW)
	require.NoError(t, err)
	defer unix.Close(raw)
	for _, length := range []int{8191, 8192, 32767, 32768, 32769, 65514, 65515} {
		innerLength := length - 24
		p := wirePacket(c)
		p = append(p, make([]byte, innerLength-32)...)
		for i := 86; i < len(p); i++ {
			p[i] = byte(i*13 + 7)
		}
		binary.BigEndian.PutUint16(p[16:18], uint16(len(p)-14))
		binary.BigEndian.PutUint16(p[38:40], uint16(length))
		binary.BigEndian.PutUint16(p[44:46], uint16(innerLength+8))
		binary.BigEndian.PutUint16(p[60:62], uint16(innerLength))
		binary.BigEndian.PutUint16(p[82:84], uint16(innerLength-20))
		p[24], p[25], p[68], p[69] = 0, 0, 0, 0
		binary.BigEndian.PutUint16(p[24:26], testpeer.Checksum(p[14:34]))
		binary.BigEndian.PutUint16(p[68:70], testpeer.Checksum(p[58:78]))
		pseudo := append([]byte(nil), p[26:34]...)
		pseudo = append(pseudo, 0, 17, p[38], p[39])
		pseudo = append(pseudo, p[34:]...)
		sum := testpeer.Checksum(pseudo)
		if sum == 0 {
			sum = 0xffff
		}
		binary.BigEndian.PutUint16(p[40:42], sum)
		corrupt := append([]byte(nil), p[14:]...)
		corrupt[len(corrupt)-1] ^= 1
		require.NoError(t, unix.Sendto(raw, corrupt, 0, &unix.SockaddrInet4{Addr: c.Local.As4()}))
		require.NoError(t, unix.Sendto(raw, p[14:], 0, &unix.SockaddrInet4{Addr: c.Local.As4()}))
		require.NoError(t, app.SetDeadline(time.Now().Add(time.Second)))
		buf := make([]byte, 65535)
		n, _, err := app.ReadFromUDP(buf)
		if err != nil {
			var counts [4]uint64
			for key := range counts {
				require.NoError(t, r.collection.Maps["counters"].Lookup(uint32(key), &counts[key]))
			}
			t.Fatalf("UDP%d reception: %v; counters UL/DL/drop/delegated %v", length, err, counts)
		}
		require.Equal(t, p[86:], buf[:n])
		t.Logf("actual completed UDP%d checksum and tail admitted via kernel", length)
	}
	ul, dl, drops, err := r.Stats()
	require.NoError(t, err)
	require.Zero(t, ul)
	require.Equal(t, uint64(7), dl)
	require.Zero(t, drops)
	require.NoError(t, app.SetReadDeadline(time.Now().Add(20*time.Millisecond)))
	_, _, err = app.ReadFromUDP(make([]byte, 65535))
	require.Error(t, err)
	require.True(t, os.IsTimeout(err), "corrupted completed tails must not reach the UE")
}
