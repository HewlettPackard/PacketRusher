//go:build linux

// SPDX-License-Identifier: Apache-2.0
// Package testpeer provides the isolated, encoded GTP peer used by native backend tests.
package testpeer

import (
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The peer runs in its own kernel network namespace so packets really cross
// Ethernet ingress/egress; the host's local-route shortcut cannot satisfy this.
func Run(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_PEER") != "1" {
		t.Skip("native helper")
	}
	barrier := os.NewFile(3, "barrier")
	ready := os.NewFile(4, "ready")
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
	// Veth TX partial checksums are not wire checksums. Make the peer produce the
	// same completed UDP checksums that a physical N3 NIC receives.
	if os.Getenv("PACKETRUSHER_PEER_KEEP_OFFLOAD") != "1" {
		out, err := exec.Command("ethtool", "-K", "pr-core", "tx", "off", "rx", "off", "gro", "off", "gso", "off", "tso", "off").CombinedOutput()
		require.NoError(t, err, string(out))
	}
	steps := strings.Split(os.Getenv("PACKETRUSHER_EBPF_PEER_STEPS"), ",")
	if steps[0] == "" {
		steps = []string{"1", "2", "1"}
	}
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.2"), Port: 2152})
	require.NoError(t, err)
	defer socket.Close()
	var targetPeer *net.UDPConn
	if strings.Contains(os.Getenv("PACKETRUSHER_EBPF_PEER_STEPS"), "-remote") {
		remote, remoteLink, cidr := "10.88.0.4", peer, "10.88.0.4/24"
		if override := os.Getenv("PACKETRUSHER_EBPF_REMOTE_ADDR"); override != "" {
			remote, remoteLink, cidr = override, lo, override+"/32"
			require.NoError(t, os.WriteFile("/proc/sys/net/ipv4/conf/all/arp_ignore", []byte("1"), 0644))
		}
		require.NoError(t, netlink.AddrAdd(remoteLink, &netlink.Addr{IPNet: Network(cidr)}))
		targetPeer, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(remote), Port: 2152})
		require.NoError(t, err)
		defer targetPeer.Close()
	}
	plain, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.2"), Port: 9999})
	require.NoError(t, err)
	defer plain.Close()
	go func() {
		b := make([]byte, 100)
		n, addr, e := plain.ReadFromUDP(b)
		if e == nil {
			_, _ = plain.WriteToUDP(b[:n], addr)
		}
	}()
	_, err = ready.Write([]byte{1})
	require.NoError(t, err)
	for _, step := range steps {
		activeSocket := socket
		if strings.HasSuffix(step, "-remote") {
			activeSocket = targetPeer
		}
		require.NoError(t, activeSocket.SetReadDeadline(time.Now().Add(10*time.Second)))
		packet := make([]byte, 1500)
		n, source, e := activeSocket.ReadFromUDP(packet)
		require.NoError(t, e)
		packet = packet[:n]
		require.GreaterOrEqual(t, n, 44)
		require.Equal(t, []byte{0x34, 255}, packet[:2])
		require.Equal(t, uint16(n-8), binary.BigEndian.Uint16(packet[2:4]))
		require.Equal(t, []byte{0, 0, 0, 0x85, 1, 0x10, 9, 0}, packet[8:16])
		ul := binary.BigEndian.Uint32(packet[4:8])
		dl := uint32(2001)
		if step == "2" || step == "2-remote" {
			require.Equal(t, uint32(1002), ul)
			require.Equal(t, "10.88.0.3", source.IP.String())
			dl = 2002
		} else {
			require.Equal(t, uint32(1001), ul)
			require.Equal(t, "10.88.0.1", source.IP.String())
		}
		// A real keepalive must traverse TCX into the joined Echo-only responder.
		t.Logf("received real uplink TEID %d from %s; checking GTP keepalive", ul, source)
		echo := []byte{0x32, 1, 0, 4, 0, 0, 0, 0, 0x12, 0x34, 0, 0}
		_, err = activeSocket.WriteToUDP(echo, source)
		require.NoError(t, err)
		require.NoError(t, activeSocket.SetReadDeadline(time.Now().Add(2*time.Second)))
		control := make([]byte, 100)
		cn, echoSource, e := activeSocket.ReadFromUDP(control)
		require.NoError(t, e)
		require.Equal(t, source, echoSource)
		require.Equal(t, []byte{0x32, 2, 0, 6, 0, 0, 0, 0, 0x12, 0x34, 0, 0, 14, 0}, control[:cn])
		inner := append([]byte(nil), packet[16:]...)
		require.Equal(t, byte(0x45), inner[0])
		require.Equal(t, []byte{10, 60, 0, 1}, inner[12:16])
		require.Equal(t, byte(17), inner[9])
		require.Equal(t, uint16(0), Checksum(inner[:20]), "uplink IPv4 checksum must be completed")
		if binary.BigEndian.Uint16(inner[26:28]) != 0 {
			pseudo := append([]byte(nil), inner[12:20]...)
			pseudo = append(pseudo, 0, 17, inner[24], inner[25])
			pseudo = append(pseudo, inner[20:]...)
			require.Equal(t, uint16(0), Checksum(pseudo), "uplink UDP checksum must be completed on the actual N3 wire")
		}
		copy(inner[12:16], packet[32:36])
		copy(inner[16:20], packet[28:32])
		inner[20], inner[21], inner[22], inner[23] = inner[22], inner[23], inner[20], inner[21]
		inner[10], inner[11], inner[26], inner[27] = 0, 0, 0, 0
		binary.BigEndian.PutUint16(inner[10:12], Checksum(inner[:20]))
		reply := append([]byte(nil), packet[:16]...)
		reply[13] = 0
		binary.BigEndian.PutUint32(reply[4:8], dl)
		reply = append(reply, inner...)
		// Wrong TEID and valid-header/wrong-UE destination must never reach the app.
		wrong := append([]byte(nil), reply...)
		binary.BigEndian.PutUint32(wrong[4:8], dl+99)
		_, err = activeSocket.WriteToUDP(wrong, source)
		require.NoError(t, err)
		wrong = append([]byte(nil), reply...)
		wrong[35] = 2
		wrong[26], wrong[27] = 0, 0
		binary.BigEndian.PutUint16(wrong[26:28], Checksum(wrong[16:36]))
		_, err = activeSocket.WriteToUDP(wrong, source)
		require.NoError(t, err)
		if step == "2-remote" {
			// The source UPF must not deliver with an otherwise current target TEID.
			_, err = socket.WriteToUDP(reply, source)
			require.NoError(t, err)
		}
		_, err = activeSocket.WriteToUDP(reply, source)
		require.NoError(t, err)
	}
}
func Network(cidr string) *net.IPNet {
	ip, prefix, e := net.ParseCIDR(cidr)
	if e != nil {
		panic(e)
	}
	prefix.IP = ip
	return prefix
}
func Checksum(b []byte) uint16 {
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

func Start(t *testing.T, helperName string, steps ...string) func() {
	t.Helper()
	barrierR, barrierW, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { barrierR.Close() })
	t.Cleanup(func() { barrierW.Close() })
	readyR, readyW, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { readyR.Close() })
	t.Cleanup(func() { readyW.Close() })
	child := exec.Command(os.Args[0], "-test.run=^"+helperName+"$", "-test.v")
	child.Env = append(os.Environ(), "PACKETRUSHER_EBPF_PEER=1")
	if len(steps) > 0 {
		child.Env = append(child.Env, "PACKETRUSHER_EBPF_PEER_STEPS="+strings.Join(steps, ","))
	}
	child.ExtraFiles = []*os.File{barrierR, readyW}
	child.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWNET}
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	require.NoError(t, child.Start())
	childDone := false
	t.Cleanup(func() {
		if !childDone {
			_ = child.Process.Kill()
			_ = child.Wait()
		}
	})
	n3 := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: "pr-n3", MTU: 1500}, PeerName: "pr-core"}
	require.NoError(t, netlink.LinkAdd(n3))
	t.Cleanup(func() { netlink.LinkDel(n3) })
	peer, err := netlink.LinkByName("pr-core")
	require.NoError(t, err)
	require.NoError(t, netlink.LinkSetNsPid(peer, child.Process.Pid))
	require.NoError(t, netlink.AddrAdd(n3, &netlink.Addr{IPNet: Network("10.88.0.1/24")}))
	require.NoError(t, netlink.AddrAdd(n3, &netlink.Addr{IPNet: Network("10.88.0.3/24")}))
	require.NoError(t, netlink.LinkSetUp(n3))
	_, err = barrierW.Write([]byte{1})
	require.NoError(t, err)
	var ready [1]byte
	require.NoError(t, readyR.SetReadDeadline(time.Now().Add(10*time.Second)))
	_, err = io.ReadFull(readyR, ready[:])
	require.NoError(t, err)
	return func() { t.Helper(); require.NoError(t, child.Wait()); childDone = true }
}
