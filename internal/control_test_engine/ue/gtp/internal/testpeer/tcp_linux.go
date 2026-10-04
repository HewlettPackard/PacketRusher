//go:build linux

// SPDX-License-Identifier: Apache-2.0
package testpeer

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

func TCPBulkPayload() []byte {
	return bytes.Repeat([]byte("PacketRusher-TCP-bulk-0123456789\n"), 8192)
}

// This UPF fixture forwards encoded GTP-U to its own DN TUN. Its actual kernel
// TCP stack, rather than hand-crafted ACKs, completes the handshake and echo.
func RunTCP(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_PEER") != "1" {
		t.Skip("native helper")
	}
	ipv6 := os.Getenv("PACKETRUSHER_TCP_FAMILY") == "6"
	dn, ue, network := "192.0.2.1", "10.60.0.1", "tcp4"
	if ipv6 {
		dn, ue, network = "2001:db8:ffff::9", "2001:db8:1234::7", "tcp6"
	}
	barrier, ready := os.NewFile(3, "barrier"), os.NewFile(4, "ready")
	defer barrier.Close()
	defer ready.Close()
	var one [1]byte
	_, err := io.ReadFull(barrier, one[:])
	require.NoError(t, err)
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	require.NoError(t, netlink.LinkSetUp(lo))
	peer, err := netlink.LinkByName("pr-core")
	require.NoError(t, err)
	require.NoError(t, netlink.AddrAdd(peer, &netlink.Addr{IPNet: Network("10.88.0.2/24")}))
	require.NoError(t, netlink.LinkSetUp(peer))
	if os.Getenv("PACKETRUSHER_TCP_KEEP_OFFLOAD") != "1" {
		out, err := exec.Command("ethtool", "-K", "pr-core", "tx", "off", "rx", "off", "gro", "off", "gso", "off", "tso", "off").CombinedOutput()
		require.NoError(t, err, string(out))
	}
	tun, device, err := userspace.NewTUN("pr-dn")
	require.NoError(t, err)
	defer tun.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 1456))
	require.NoError(t, netlink.AddrAdd(device, &netlink.Addr{IPNet: hostNetwork(dn), Flags: unix.IFA_F_NODAD}))
	require.NoError(t, netlink.LinkSetUp(device))
	require.NoError(t, netlink.RouteAdd(&netlink.Route{Dst: hostNetwork(ue), LinkIndex: device.Attrs().Index, Scope: netlink.SCOPE_LINK}))
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.2"), Port: 2152})
	require.NoError(t, err)
	defer udp.Close()
	listener, err := net.ListenTCP(network, &net.TCPAddr{IP: net.ParseIP(dn), Port: 9000})
	require.NoError(t, err)
	defer listener.Close()
	require.NoError(t, listener.SetDeadline(time.Now().Add(10*time.Second)))
	stopped := make(chan struct{})
	results := make(chan error, 2)
	closed := func() bool {
		select {
		case <-stopped:
			return true
		default:
			return false
		}
	}
	go func() {
		packet := make([]byte, 1500)
		for {
			n, source, err := udp.ReadFromUDP(packet)
			if err != nil {
				if closed() {
					err = nil
				}
				results <- err
				return
			}
			if source.IP.String() != "10.88.0.1" || source.Port != 2152 || n < 56 || packet[0] != 0x34 || packet[1] != 255 || binary.BigEndian.Uint16(packet[2:4]) != uint16(n-8) || binary.BigEndian.Uint32(packet[4:8]) != 1001 || !bytes.Equal(packet[8:16], []byte{0, 0, 0, 0x85, 1, 0x10, 9, 0}) {
				results <- fmt.Errorf("unexpected uplink tuple/header from %s: %x", source, packet[:n])
				return
			}
			inner := packet[16:n]
			if ipv6 && len(inner) == 48 && inner[0]>>4 == 6 && inner[6] == 58 && inner[40] == 133 {
				ra, _ := hex.DecodeString("6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000")
				gtp, _ := userspace.Encode(2001, 9, ra)
				gtp[13] = 0
				if _, err := udp.WriteToUDP(gtp, source); err != nil {
					results <- err
					return
				}
				continue
			}
			if err := validateTCPPacket(inner, ue, dn, ipv6); err != nil {
				t.Logf("actual uplink admission error: %v", err)
				results <- err
				return
			}

			if _, err := tun.Write(packet[16:n]); err != nil {
				if closed() {
					err = nil
				}
				results <- err
				return
			}
		}
	}()
	go func() {
		packet := make([]byte, 1500)
		for {
			n, err := tun.Read(packet)
			if err != nil {
				if closed() {
					err = nil
				}
				results <- err
				return
			}
			// The private DN TUN starts IPv6 router discovery independently of
			// this IPv4 fixture. It is not a downlink PDU and must not stop its
			// IPv4 forwarding worker. Other unexpected packets still fail.
			if n == 48 && packet[0]>>4 == 6 && packet[6] == 58 && packet[7] == 255 && packet[40] == 133 && packet[41] == 0 {
				continue
			}
			if err := validateTCPPacket(packet[:n], dn, ue, ipv6); err != nil {
				t.Logf("actual downlink admission error: %v", err)
				results <- err
				return
			}

			gtp := []byte{0x34, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x85, 1, 0, 9, 0}
			binary.BigEndian.PutUint16(gtp[2:4], uint16(n+8))
			binary.BigEndian.PutUint32(gtp[4:8], 2001)
			_, err = udp.WriteToUDP(append(gtp, packet[:n]...), &net.UDPAddr{IP: net.ParseIP("10.88.0.1"), Port: 2152})
			if err != nil {
				if closed() {
					err = nil
				}
				results <- err
				return
			}
		}
	}()
	defer func() {
		close(stopped)
		_ = udp.Close()
		_ = tun.Close()
		for i := 0; i < 2; i++ {
			select {
			case err := <-results:
				require.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("owned fake UPF transport worker did not join")
			}
		}
	}()
	_, err = ready.Write([]byte{1})
	require.NoError(t, err)
	conn, err := listener.AcceptTCP()
	require.NoError(t, err)
	defer conn.Close()
	require.NoError(t, conn.SetDeadline(time.Now().Add(10*time.Second)))
	packet := make([]byte, len(TCPBulkPayload()))
	_, err = io.ReadFull(conn, packet)
	require.NoError(t, err)
	require.Equal(t, TCPBulkPayload(), packet)
	_, err = io.CopyN(conn, bytes.NewReader(packet), int64(len(packet)))
	require.NoError(t, err)
	// Write completion only queues TCP data. Keep the UPF/TUN alive until the
	// application confirms it received the full echo; tearing it down here
	// would discard queued tail segments and their acknowledgments.
	_, err = io.ReadFull(conn, one[:])
	require.NoError(t, err)
	require.Equal(t, byte(1), one[0])
}

func hostNetwork(address string) *net.IPNet {
	ip := net.ParseIP(address)
	if ip.To4() != nil {
		return &net.IPNet{IP: ip.To4(), Mask: net.CIDRMask(32, 32)}
	}
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
}
func validateTCPPacket(packet []byte, source, destination string, ipv6 bool) error {
	var pseudo []byte
	var tcp []byte
	if ipv6 {
		if len(packet) < 60 || packet[0]>>4 != 6 || packet[6] != 6 || int(binary.BigEndian.Uint16(packet[4:6]))+40 != len(packet) || !net.IP(packet[8:24]).Equal(net.ParseIP(source)) || !net.IP(packet[24:40]).Equal(net.ParseIP(destination)) {
			return fmt.Errorf("unexpected IPv6 TCP packet: %x", packet)
		}
		pseudo = append(pseudo, packet[8:40]...)
		pseudo = binary.BigEndian.AppendUint32(pseudo, uint32(len(packet)-40))
		pseudo = append(pseudo, 0, 0, 0, 6)
		tcp = packet[40:]
	} else {
		if len(packet) < 40 || packet[0] != 0x45 || packet[9] != 6 || int(binary.BigEndian.Uint16(packet[2:4])) != len(packet) || Checksum(packet[:20]) != 0 || !net.IP(packet[12:16]).Equal(net.ParseIP(source)) || !net.IP(packet[16:20]).Equal(net.ParseIP(destination)) {
			return fmt.Errorf("unexpected IPv4 TCP packet: %x", packet)
		}
		pseudo = append(pseudo, packet[12:20]...)
		pseudo = append(pseudo, 0, 6, byte((len(packet)-20)>>8), byte(len(packet)-20))
		tcp = packet[20:]
	}
	pseudo = append(pseudo, tcp...)
	if sum := Checksum(pseudo); sum != 0 {
		return fmt.Errorf("invalid TCP checksum on real N3 wire: %04x", sum)
	}
	return nil
}
