// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"
)

// PacketPort transfers one raw IP packet per read or write (Linux IFF_NO_PI TUN).
type PacketPort interface{ io.ReadWriteCloser }

type Config struct {
	Local                    netip.Addr
	Remote                   netip.AddrPort
	UplinkTEID, DownlinkTEID uint32
	QFI                      uint8
	IPv4                     netip.Addr
	AllowIPv6                bool
	IPv6PrefixAllowed        func(netip.Addr) bool
	IPv6InterfaceID          [8]byte
}

type binding struct {
	session *Session
	config  Config
}
type endpoint struct {
	conn     *net.UDPConn
	mu       sync.RWMutex
	sessions map[uint32]binding
	refs     int
	done     chan struct{}
}

type Registry struct {
	mu        sync.Mutex
	endpoints map[netip.Addr]*endpoint
}

func NewRegistry() *Registry { return &Registry{endpoints: make(map[netip.Addr]*endpoint)} }

// A registry shares UDP2152 across UEs but has no reuse-port or promiscuous sockets.
// Each endpoint exists only while sessions hold it, and final release waits for its
// receive worker so a later registration can bind the same N3 address immediately.
var DefaultRegistry = NewRegistry()

func validate(c Config) error {
	if !c.Local.Is4() || !c.Remote.IsValid() || !c.Remote.Addr().Is4() || c.Remote.Port() == 0 || c.UplinkTEID == 0 || c.DownlinkTEID == 0 || c.QFI > 63 {
		return errors.New("invalid userspace GTP-U endpoint, TEID or QFI")
	}
	if !c.IPv4.Is4() && !c.AllowIPv6 {
		return errors.New("session has no IPv4 or IPv6 address")
	}
	return nil
}
func (r *Registry) take(c Config, s *Session) (*endpoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.endpoints[c.Local]
	if e == nil {
		conn, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(c.Local, 2152)))
		if err != nil {
			return nil, fmt.Errorf("bind userspace N3 %s:2152: %w", c.Local, err)
		}
		e = &endpoint{conn: conn, sessions: make(map[uint32]binding), done: make(chan struct{})}
		r.endpoints[c.Local] = e
		go e.receive()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if existing, ok := e.sessions[c.DownlinkTEID]; ok && existing.session != s {
		return nil, fmt.Errorf("downlink TEID %d already belongs to another session", c.DownlinkTEID)
	}
	e.sessions[c.DownlinkTEID] = binding{s, c}
	e.refs++
	return e, nil
}
func (r *Registry) give(e *endpoint, c Config, s *Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.mu.Lock()
	if held, ok := e.sessions[c.DownlinkTEID]; ok && held.session == s {
		delete(e.sessions, c.DownlinkTEID)
	}
	e.refs--
	e.mu.Unlock()
	if e.refs == 0 {
		e.conn.Close()
		<-e.done
		delete(r.endpoints, c.Local)
	}
}
func (e *endpoint) receive() {
	defer close(e.done)
	buf := make([]byte, 65535)
	for {
		n, peer, err := e.conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			return
		}
		if response := echoResponse(buf[:n]); response != nil {
			_, _ = e.conn.WriteToUDPAddrPort(response, peer)
			continue
		}
		teid, payload, err := Decode(buf[:n])
		if err != nil {
			continue
		}
		e.mu.RLock()
		held, ok := e.sessions[teid]
		e.mu.RUnlock()
		if !ok || peer != held.config.Remote {
			continue
		}
		_, dst, err := IPAddresses(payload)
		if err != nil {
			continue
		}
		addr, ok := netip.AddrFromSlice(dst)
		if !ok || (addr.Is4() && addr != held.config.IPv4) || (addr.Is6() && (!held.config.AllowIPv6 || (!held.config.allowsIPv6(addr) && !(addr.IsMulticast() && routerAdvertisement(payload))))) {
			continue
		}
		held.session.deliver(payload)
	}
}

type Session struct {
	mu             sync.RWMutex
	registry       *Registry
	port           PacketPort
	endpoint       *endpoint
	config         Config
	stopped        bool
	once           sync.Once
	done           chan struct{}
	advertisements chan []byte
	downlink       chan []byte
	stop           chan struct{}
	received       chan struct{}
}

func (r *Registry) Open(port PacketPort, c Config) (*Session, error) {
	if port == nil {
		return nil, errors.New("packet port is missing")
	}
	if err := validate(c); err != nil {
		return nil, err
	}
	s := &Session{registry: r, port: port, config: c, done: make(chan struct{}), advertisements: make(chan []byte, 8), downlink: make(chan []byte, 64), stop: make(chan struct{}), received: make(chan struct{})}
	e, err := r.take(c, s)
	if err != nil {
		return nil, err
	}
	s.endpoint = e
	go s.uplink()
	go s.downlinkWorker()
	return s, nil
}

// Update stages the new endpoint and TEID before replacing the active sender.
// Session close and concurrent packet sends cannot race with socket retirement.
func (s *Session) Update(c Config) error {
	if err := validate(c); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return net.ErrClosed
	}
	oldE, oldC := s.endpoint, s.config
	if oldC.Local == c.Local && oldC.DownlinkTEID == c.DownlinkTEID {
		oldE.mu.Lock()
		oldE.sessions[c.DownlinkTEID] = binding{s, c}
		oldE.mu.Unlock()
		s.config = c
		return nil
	}
	e, err := s.registry.take(c, s)
	if err != nil {
		return err
	}
	s.endpoint, s.config = e, c
	s.registry.give(oldE, oldC, s)
	return nil
}

// IPv6 NAS allocates only the IID. Global traffic is allowed after a validated
// RA supplies the prefix, and stops immediately when that allocation expires.
func (c Config) allowsIPv6(address netip.Addr) bool {
	bytes16 := address.As16()
	if !bytes.Equal(bytes16[8:], c.IPv6InterfaceID[:]) {
		return false
	}
	if address.IsLinkLocalUnicast() {
		return true
	}
	return address.IsGlobalUnicast() && c.IPv6PrefixAllowed != nil && c.IPv6PrefixAllowed(address)
}

func (s *Session) Send(payload []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.stopped {
		return net.ErrClosed
	}
	src, _, err := IPAddresses(payload)
	if err != nil {
		return err
	}
	addr, ok := netip.AddrFromSlice(src)
	if !ok || (addr.Is4() && addr != s.config.IPv4) || (addr.Is6() && (!s.config.AllowIPv6 || !s.config.allowsIPv6(addr))) {
		return errors.New("packet source does not belong to session")
	}
	b, err := Encode(s.config.UplinkTEID, s.config.QFI, payload)
	if err != nil {
		return err
	}
	_, err = s.endpoint.conn.WriteToUDPAddrPort(b, s.config.Remote)
	return err
}
func (s *Session) uplink() {
	defer close(s.done)
	buf := make([]byte, 65535)
	for {
		n, err := s.port.Read(buf)
		if err != nil {
			return
		}
		_ = s.Send(buf[:n])
	}
}
func (s *Session) deliver(payload []byte) {
	// A slow UE cannot stall other UEs sharing the N3 receiver. Queues have a fixed
	// capacity, and packets are copied before the endpoint reuses its UDP buffer.
	select {
	case <-s.stop:
		return
	default:
	}
	copied := append([]byte(nil), payload...)
	select {
	case s.downlink <- copied:
	case <-s.stop:
	default:
	}
}
func (s *Session) downlinkWorker() {
	defer close(s.received)
	for {
		select {
		case <-s.stop:
			return
		case payload := <-s.downlink:
			if routerAdvertisement(payload) {
				select {
				case s.advertisements <- append([]byte(nil), payload...):
				default:
				}
				continue // Prefix discovery owns RA; prevent host SLAAC/default-route side effects.
			}
			_, _ = s.port.Write(payload)
		}
	}
}
func (s *Session) Advertisements() <-chan []byte { return s.advertisements }
func (s *Session) Close() {
	s.once.Do(func() {
		s.mu.Lock()
		s.stopped = true
		e, c := s.endpoint, s.config
		s.mu.Unlock()
		close(s.stop)
		s.port.Close() // Interrupt TUN read/write before waiting for workers.
		<-s.done
		<-s.received
		s.registry.give(e, c, s)
	})
}
