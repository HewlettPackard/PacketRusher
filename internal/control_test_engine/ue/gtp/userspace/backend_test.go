// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"
)

type packetPipe struct {
	read, write chan []byte
	closed      chan struct{}
	once        sync.Once
}

func newPacketPipe() *packetPipe {
	return &packetPipe{read: make(chan []byte, 128), write: make(chan []byte, 128), closed: make(chan struct{})}
}
func (p *packetPipe) Read(b []byte) (int, error) {
	select {
	case <-p.closed:
		return 0, io.EOF
	case packet := <-p.read:
		return copy(b, packet), nil
	}
}
func (p *packetPipe) Write(b []byte) (int, error) {
	select {
	case <-p.closed:
		return 0, io.EOF
	case p.write <- append([]byte(nil), b...):
		return len(b), nil
	}
}
func (p *packetPipe) Close() error { p.once.Do(func() { close(p.closed) }); return nil }
func udpPeer(t *testing.T) *net.UDPConn {
	t.Helper()
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 0, 9)})
	require.NoError(t, err)
	t.Cleanup(func() { c.Close() })
	return c
}
func packetFrom(t *testing.T, c *net.UDPConn) ([]byte, netip.AddrPort) {
	t.Helper()
	require.NoError(t, c.SetReadDeadline(time.Now().Add(time.Second)))
	buf := make([]byte, 65535)
	n, peer, err := c.ReadFromUDPAddrPort(buf)
	require.NoError(t, err)
	return buf[:n], peer
}
func downlink(t *testing.T, c *net.UDPConn, local netip.Addr, teid uint32, ip []byte) {
	t.Helper()
	p, err := Encode(teid, 9, ip)
	require.NoError(t, err)
	_, err = c.WriteToUDPAddrPort(p, netip.AddrPortFrom(local, 2152))
	require.NoError(t, err)
}
func receiveInner(t *testing.T, p *packetPipe) []byte {
	t.Helper()
	select {
	case b := <-p.write:
		return b
	case <-time.After(time.Second):
		t.Fatal("no decapsulated packet")
		return nil
	}
}
func testConfig(local string, peer *net.UDPConn, id uint32, ue string) Config {
	return Config{Local: netip.MustParseAddr(local), Remote: peer.LocalAddr().(*net.UDPAddr).AddrPort(), UplinkTEID: id + 100, DownlinkTEID: id, QFI: 9, IPv4: netip.MustParseAddr(ue)}
}
func TestSharedSocketIsolationHandoverAndRelease(t *testing.T) {
	peer := udpPeer(t)
	registry := NewRegistry()
	a, b := newPacketPipe(), newPacketPipe()
	ca := testConfig("127.88.0.1", peer, 1, "10.0.0.1")
	cb := testConfig("127.88.0.1", peer, 2, "10.0.0.2")
	sa, err := registry.Open(a, ca)
	require.NoError(t, err)
	defer sa.Close()
	sb, err := registry.Open(b, cb)
	require.NoError(t, err)
	defer sb.Close()
	require.Len(t, registry.endpoints, 1)
	packet := ipv4([4]byte{10, 0, 0, 1}, [4]byte{8, 8, 8, 8})
	a.read <- packet
	wire, source := packetFrom(t, peer)
	id, inner, err := Decode(wire)
	require.NoError(t, err)
	require.Equal(t, ca.UplinkTEID, id)
	require.Equal(t, packet, inner)
	require.Equal(t, netip.AddrPortFrom(ca.Local, 2152), source)
	require.Equal(t, byte(9), wire[14])
	for _, entry := range []struct {
		cfg  Config
		port *packetPipe
		dst  [4]byte
	}{{ca, a, [4]byte{10, 0, 0, 1}}, {cb, b, [4]byte{10, 0, 0, 2}}} {
		payload := ipv4([4]byte{8, 8, 8, 8}, entry.dst)
		downlink(t, peer, entry.cfg.Local, entry.cfg.DownlinkTEID, payload)
		require.Equal(t, payload, receiveInner(t, entry.port))
	}
	// Invalid target ownership must leave the old binding fully usable.
	conflict := ca
	conflict.DownlinkTEID = cb.DownlinkTEID
	require.Error(t, sa.Update(conflict))
	require.NoError(t, sa.Send(packet))
	wire, _ = packetFrom(t, peer)
	id, _, err = Decode(wire)
	require.NoError(t, err)
	require.Equal(t, ca.UplinkTEID, id)
	target := ca
	target.Local = netip.MustParseAddr("127.88.0.2")
	target.DownlinkTEID = 3
	target.UplinkTEID = 103
	target.QFI = 63
	require.NoError(t, sa.Update(target))
	require.Len(t, registry.endpoints, 2)
	require.NoError(t, sa.Send(packet))
	wire, source = packetFrom(t, peer)
	id, _, err = Decode(wire)
	require.NoError(t, err)
	require.Equal(t, target.UplinkTEID, id)
	require.Equal(t, target.Local, source.Addr())
	require.Equal(t, byte(63), wire[14])
	payload := ipv4([4]byte{8, 8, 8, 8}, [4]byte{10, 0, 0, 1})
	downlink(t, peer, target.Local, target.DownlinkTEID, payload)
	require.Equal(t, payload, receiveInner(t, a))
	sa.Close()
	sa.Close()
	require.Len(t, registry.endpoints, 1)
	// Rebinding source after final target release proves worker/socket cleanup.
	rebound, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(target.Local, 2152)))
	require.NoError(t, err)
	rebound.Close()
	payload = ipv4([4]byte{8, 8, 8, 8}, [4]byte{10, 0, 0, 2})
	downlink(t, peer, cb.Local, cb.DownlinkTEID, payload)
	require.Equal(t, payload, receiveInner(t, b))
	sb.Close()
	require.Empty(t, registry.endpoints)
}
func TestSpoofedPeerWrongTEIDAndDestinationAreDropped(t *testing.T) {
	peer := udpPeer(t)
	other := udpPeer(t)
	registry := NewRegistry()
	pipe := newPacketPipe()
	cfg := testConfig("127.88.1.1", peer, 8, "10.0.0.8")
	s, err := registry.Open(pipe, cfg)
	require.NoError(t, err)
	defer s.Close()
	correct := ipv4([4]byte{8, 8, 8, 8}, [4]byte{10, 0, 0, 8})
	downlink(t, other, cfg.Local, 8, correct)
	downlink(t, peer, cfg.Local, 9, correct)
	downlink(t, peer, cfg.Local, 8, ipv4([4]byte{8, 8, 8, 8}, [4]byte{10, 0, 0, 9}))
	select {
	case <-pipe.write:
		t.Fatal("unauthorized packet reached UE")
	case <-time.After(50 * time.Millisecond):
	}
	downlink(t, peer, cfg.Local, 8, correct)
	require.Equal(t, correct, receiveInner(t, pipe))
	require.Error(t, s.Send(ipv4([4]byte{10, 0, 0, 9}, [4]byte{8, 8, 8, 8})))
}
func TestConcurrentUpdateSendAndClose(t *testing.T) {
	peer := udpPeer(t)
	r := NewRegistry()
	cfg := testConfig("127.88.2.1", peer, 11, "10.0.0.11")
	s, err := r.Open(newPacketPipe(), cfg)
	require.NoError(t, err)
	packet := ipv4([4]byte{10, 0, 0, 11}, [4]byte{8, 8, 8, 8})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if worker == 0 {
					next := cfg
					next.UplinkTEID += uint32(j)
					next.DownlinkTEID += uint32(j)
					_ = s.Update(next)
				} else {
					_ = s.Send(packet)
				}
			}
		}(i)
	}
	wg.Add(1)
	go func() { defer wg.Done(); s.Close() }()
	wg.Wait()
	require.Empty(t, r.endpoints)
	require.ErrorIs(t, s.Send(packet), net.ErrClosed)
	require.ErrorIs(t, s.Update(cfg), net.ErrClosed)
}
