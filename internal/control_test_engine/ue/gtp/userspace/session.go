/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package userspace is a GTP-U datapath that needs no kernel module: packets move
// between one TUN device per UE and one UDP socket per gNB N3 address.
package userspace

import (
	"errors"
	"net"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// Config is what a session needs from its PDU session.
type Config struct {
	UPF                      netip.Addr
	UplinkTEID, DownlinkTEID uint32
	QFI                      uint8
	UE                       netip.Addr // packets from or to another address are dropped
}

// Session carries the packets of one UE between its TUN device and the UPF.
type Session struct {
	name     string
	tun      *os.File
	endpoint *endpoint
	config   atomic.Pointer[Config]
	done     chan struct{}
	warned   sync.Once
}

// endpoint is the socket of one N3 address, shared by the UEs of that gNB.
type endpoint struct {
	local    netip.Addr
	conn     *net.UDPConn
	done     chan struct{}
	mu       sync.RWMutex
	sessions map[uint32]*Session // by downlink TEID
}

var (
	mu        sync.Mutex
	endpoints = map[netip.Addr]*endpoint{}
)

// Open creates the TUN device name and carries its packets over GTP-U between the
// N3 address local and the UPF.
func Open(name string, local netip.Addr, c Config) (*Session, error) {
	// The device goes away with its file, and no other process may attach to it.
	tun := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: name}, Mode: netlink.TUNTAP_MODE_TUN,
		Flags: netlink.TUNTAP_NO_PI | netlink.TUNTAP_TUN_EXCL, Queues: 1, NonPersist: true}
	if err := netlink.LinkAdd(tun); err != nil {
		return nil, err
	}
	s := &Session{name: name, tun: tun.Fds[0], done: make(chan struct{})}
	s.config.Store(&c)
	if err := netlink.LinkSetUp(tun); err != nil {
		s.tun.Close()
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	if s.endpoint = endpoints[local]; s.endpoint == nil {
		conn, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(local, 2152)))
		if err != nil {
			s.tun.Close()
			return nil, err
		}
		s.endpoint = &endpoint{local: local, conn: conn, done: make(chan struct{}), sessions: map[uint32]*Session{}}
		endpoints[local] = s.endpoint
		go s.endpoint.receive()
	}
	s.Update(c)
	go s.uplink()
	return s, nil
}

// Update applies the TEIDs, UPF and QFI of the PDU session on the same N3 address.
func (s *Session) Update(c Config) {
	s.endpoint.mu.Lock()
	defer s.endpoint.mu.Unlock()
	delete(s.endpoint.sessions, s.config.Load().DownlinkTEID)
	s.endpoint.sessions[c.DownlinkTEID] = s
	s.config.Store(&c)
}

// Close removes the TUN device and, with the last session of its N3 address, the
// socket: the address can be bound again once Close returns.
func (s *Session) Close() {
	s.tun.Close()
	<-s.done
	e := s.endpoint
	mu.Lock()
	defer mu.Unlock()
	e.mu.Lock()
	delete(e.sessions, s.config.Load().DownlinkTEID)
	last := len(e.sessions) == 0
	e.mu.Unlock()
	if last {
		delete(endpoints, e.local)
		e.conn.Close()
		<-e.done
	}
}

// uplink sends what the UE's applications route into the TUN device to the UPF.
// The GTP-U header is written in front of each packet, in the same buffer.
func (s *Session) uplink() {
	defer close(s.done)
	buf := make([]byte, Headroom+65535)
	for {
		n, err := s.tun.Read(buf[Headroom:])
		if err != nil {
			s.warn("reading uplink packets", err)
			return
		}
		c := s.config.Load()
		if !belongs(buf[Headroom:Headroom+n], 12, c.UE) {
			continue
		}
		datagram := Encode(buf[:Headroom+n], c.UplinkTEID, c.QFI)
		if _, err := s.endpoint.conn.WriteToUDPAddrPort(datagram, netip.AddrPortFrom(c.UPF, 2152)); err != nil {
			s.warn("sending to the UPF", err)
		}
	}
}

// receive writes the packets the UPF sends to this N3 address straight to the TUN
// device of the session that owns their TEID.
func (e *endpoint) receive() {
	defer close(e.done)
	buf := make([]byte, 65535)
	for {
		n, peer, err := e.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Warn("[UE][GTP] N3 socket of ", e.local, " stopped receiving: ", err)
			}
			return
		}
		if response := echoResponse(buf[:n]); response != nil {
			_, _ = e.conn.WriteToUDPAddrPort(response, peer)
			continue
		}
		teid, packet, err := Decode(buf[:n])
		if err != nil {
			continue
		}
		e.mu.RLock()
		s := e.sessions[teid]
		e.mu.RUnlock()
		if s == nil {
			continue
		}
		// The UPF may send from any UDP port (TS 29.281 §4.4.2): compare its address only.
		if c := s.config.Load(); peer.Addr() != c.UPF || !belongs(packet, 16, c.UE) {
			continue
		}
		if _, err := s.tun.Write(packet); err != nil {
			s.warn("delivering downlink packets", err)
		}
	}
}

// belongs reports whether the IPv4 packet has the UE's address at offset: 12 for
// its source, 16 for its destination.
func belongs(packet []byte, offset int, ue netip.Addr) bool {
	return len(packet) >= 20 && packet[0]>>4 == 4 && netip.AddrFrom4([4]byte(packet[offset:offset+4])) == ue
}

// warn logs the first datapath error of a session that is not being closed.
func (s *Session) warn(action string, err error) {
	if !errors.Is(err, os.ErrClosed) && !errors.Is(err, net.ErrClosed) {
		s.warned.Do(func() { log.Warn("[UE][GTP] ", s.name, ": ", action, ": ", err) })
	}
}
