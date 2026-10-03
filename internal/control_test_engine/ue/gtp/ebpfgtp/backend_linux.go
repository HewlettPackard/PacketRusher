//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
)

//go:embed gtpu_bpfel.o
var object []byte

// Config binds one IPv4 PDU session to an owned, stable L3 endpoint.
// This backend supports Ethernet N3, policy routing and an MTU <= 1500.
type Config struct {
	Local, Remote, IPv4      netip.Addr
	UplinkTEID, DownlinkTEID uint32
	QFI                      uint8
	EndpointIfIndex, MTU     int
}

func (c Config) Validate() error {
	for _, a := range []netip.Addr{c.Local, c.Remote, c.IPv4} {
		if !a.Is4() || a.IsUnspecified() || a.IsMulticast() || a.IsLoopback() {
			return errors.New("eBPF requires unicast IPv4 UE and N3 addresses")
		}
	}
	if c.Local == c.Remote || c.IPv4 == c.Local || c.IPv4 == c.Remote {
		return errors.New("eBPF UE, local N3 and peer addresses must be distinct")
	}
	if c.UplinkTEID == 0 || c.DownlinkTEID == 0 || c.QFI > 63 || c.EndpointIfIndex <= 0 || c.MTU < 68 || c.MTU > 1456 {
		return errors.New("invalid eBPF TEID, QFI, endpoint or IPv4 MTU (68..1456)")
	}
	return nil
}

type binding struct {
	Local, Peer, UE, UplinkTEID, DownlinkTEID, Endpoint, MTU uint32
	QFI                                                      uint8
	Padding                                                  [3]byte
}
type downKey struct{ Local, Peer, TEID uint32 }

func ipv4(a netip.Addr) uint32 { b := a.As4(); return binary.LittleEndian.Uint32(b[:]) }
func (c Config) binding() binding {
	return binding{Local: ipv4(c.Local), Peer: ipv4(c.Remote), UE: ipv4(c.IPv4), UplinkTEID: c.UplinkTEID, DownlinkTEID: c.DownlinkTEID, Endpoint: uint32(c.EndpointIfIndex), MTU: uint32(c.MTU), QFI: c.QFI}
}
func (c Config) downKey() downKey { return downKey{ipv4(c.Local), ipv4(c.Remote), c.DownlinkTEID} }

type store interface {
	put(string, any, any, ebpf.MapUpdateFlags) error
	del(string, any) error
}
type kernelStore struct{ maps map[string]*ebpf.Map }

func (s kernelStore) put(name string, k, v any, flags ebpf.MapUpdateFlags) error {
	return s.maps[name].Update(k, v, flags)
}
func (s kernelStore) del(name string, k any) error {
	err := s.maps[name].Delete(k)
	if errors.Is(err, ebpf.ErrKeyNotExist) {
		return nil
	}
	return err
}

type closer interface{ Close() error }
type n3Lease struct {
	refs           int
	ifindex        int
	socket         closer
	mapped, linked bool
}
type ingress struct {
	refs       int
	attachment closer
}

// Registry owns only its BPF objects, TCX links and exclusively bound N3 ports.
// Its socket worker answers only owned Echo Requests; user-plane packets are
// encapsulated and decapsulated entirely by the kernel programs.
type Registry struct {
	peers           atomic.Pointer[map[peerKey]bool]
	mu              sync.Mutex
	collection      *ebpf.Collection
	state           store
	sessions        map[uint32]*Session
	orphans         []*Session
	ports           map[netip.Addr]closer
	locals          map[netip.Addr]*n3Lease
	links           map[int]*ingress
	discover        func(Config) (int, error)
	listen          func(netip.Addr) (closer, error)
	attach          func(int) (closer, error)
	warning         func(error)
	pendingWarnings []error
}

func NewRegistry() *Registry {
	r := &Registry{ports: make(map[netip.Addr]closer), sessions: make(map[uint32]*Session), locals: make(map[netip.Addr]*n3Lease), links: make(map[int]*ingress), discover: discoverN3, warning: func(error) {}}
	r.listen = func(a netip.Addr) (closer, error) {
		return newManagementSocket(a, func(peer netip.AddrPort) bool {
			snapshot := r.peers.Load()
			return peer.Port() == 2152 && snapshot != nil && (*snapshot)[peerKey{a, peer.Addr()}]
		})
	}
	r.attach = func(index int) (closer, error) {
		return link.AttachTCX(link.TCXOptions{Interface: index, Program: r.collection.Programs["decap"], Attach: ebpf.AttachTCXIngress, Anchor: link.Tail()})
	}
	return r
}

var DefaultRegistry = NewRegistry()

// SetWarningHandler observes failed post-commit retirement without misreporting
// the new atomic mapping as an unsuccessful handover.
func (r *Registry) SetWarningHandler(fn func(error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fn == nil {
		fn = func(error) {}
	}
	r.warning = fn
}

// unlock delivers observations after committing ownership and releasing mu.
// Handlers may inspect the registry without deadlocking setup or release.
func (r *Registry) unlock() {
	warnings := r.pendingWarnings
	r.pendingWarnings = nil
	handler := r.warning
	r.mu.Unlock()
	for _, err := range warnings {
		handler(err)
	}
}
func (r *Registry) queueWarning(err error) { r.pendingWarnings = append(r.pendingWarnings, err) }
func (r *Registry) load() error {
	if r.state != nil {
		return nil
	}
	switch runtime.GOARCH {
	case "amd64", "arm64", "386", "riscv64":
	default:
		return errors.New("embedded eBPF object requires a supported little-endian architecture")
	}
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(object))
	if err != nil {
		return err
	}
	spec.Programs["decap"].AttachType = ebpf.AttachTCXIngress
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("load PacketRusher eBPF programs (CAP_BPF/CAP_NET_ADMIN and TCX/LWT support required): %w", err)
	}
	r.collection = collection
	r.state = kernelStore{collection.Maps}
	return nil
}
func discoverN3(c Config) (int, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return 0, err
	}
	index := 0
	for _, l := range links {
		addresses, err := netlink.AddrList(l, netlink.FAMILY_V4)
		if err != nil {
			return 0, err
		}
		for _, a := range addresses {
			if a.IP.Equal(net.IP(c.Local.AsSlice())) {
				if index != 0 {
					return 0, errors.New("eBPF does not support an N3 address assigned to multiple interfaces")
				}
				at := l.Attrs()
				if (l.Type() != "device" && l.Type() != "veth") || at.EncapType != "ether" || len(at.HardwareAddr) != 6 || at.MasterIndex != 0 || at.Flags&net.FlagUp == 0 || at.MTU > 1500 || c.MTU+44 > at.MTU {
					return 0, errors.New("eBPF requires an up Ethernet N3 interface without VRF/master, MTU <=1500 and sufficient tunnel headroom")
				}
				index = at.Index
			}
		}
	}
	if index == 0 {
		return 0, fmt.Errorf("eBPF N3 address %s is not assigned", c.Local)
	}
	routes, err := netlink.RouteGetWithOptions(net.IP(c.Remote.AsSlice()), &netlink.RouteGetOptions{SrcAddr: net.IP(c.Local.AsSlice())})
	if err != nil {
		return 0, err
	}
	if len(routes) != 1 || routes[0].LinkIndex != index || len(routes[0].MultiPath) != 0 {
		return 0, errors.New("eBPF requires a single N3 route through the assigned Ethernet interface")
	}
	return index, nil
}
func (r *Registry) acquire(local netip.Addr, index int) error {
	if r.ports[local] != nil {
		return errors.New("N3 port cleanup is pending")
	}
	if n := r.locals[local]; n != nil {
		if !n.mapped || !n.linked {
			return errors.New("N3 lease cleanup is pending")
		}
		if n.ifindex != index {
			return errors.New("N3 interface changed while the address is in use")
		}
		n.refs++
		return nil
	}
	socket, err := r.listen(local)
	if err != nil {
		return fmt.Errorf("reserve eBPF N3 %s:2152: %w", local, err)
	}
	in := r.links[index]
	if in == nil {
		attachment, err := r.attach(index)
		if err != nil {
			if closeErr := socket.Close(); closeErr != nil {
				r.ports[local] = socket
				err = errors.Join(err, closeErr)
			}
			return fmt.Errorf("attach eBPF TCX ingress: %w", err)
		}
		in = &ingress{attachment: attachment}
		r.links[index] = in
	}
	if err = r.state.put("locals", ipv4(local), uint32(index), ebpf.UpdateNoExist); err != nil {
		if in.refs == 0 {
			if closeErr := in.attachment.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			} else {
				delete(r.links, index)
			}
		}
		if closeErr := socket.Close(); closeErr != nil {
			r.ports[local] = socket
			err = errors.Join(err, closeErr)
		}
		return err
	}
	in.refs++
	r.locals[local] = &n3Lease{refs: 1, ifindex: index, socket: socket, mapped: true, linked: true}
	return nil
}
func (r *Registry) release(local netip.Addr) error {
	n := r.locals[local]
	if n == nil {
		return nil
	}
	if n.refs > 1 {
		n.refs--
		return nil
	}
	// Track each completed phase. Failed close/detach remains reachable, and a
	// later Close must not decrement another session's shared ingress reference.
	if n.mapped {
		if err := r.state.del("locals", ipv4(local)); err != nil {
			return err
		}
		n.mapped = false
	}
	if n.linked {
		in := r.links[n.ifindex]
		if in.refs == 1 {
			if err := in.attachment.Close(); err != nil {
				return err
			}
			delete(r.links, n.ifindex)
		} else {
			in.refs--
		}
		n.linked = false
	}
	if err := n.socket.Close(); err != nil {
		return err
	}
	delete(r.locals, local)
	return nil
}

// Session keeps the same endpoint and route program across TEID/N3 handover.
type Session struct {
	registry *Registry
	cfg      Config
	keys     map[downKey]bool
	leases   map[netip.Addr]bool
	closed   bool
	stopping bool
}

func (r *Registry) Open(c Config) (*Session, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.unlock()
	if r.sessions[ipv4(c.IPv4)] != nil {
		return nil, errors.New("eBPF UE IPv4 address already has a session")
	}
	r.reap()
	index, err := r.discover(c)
	if err != nil {
		return nil, err
	}
	if err = r.load(); err != nil {
		return nil, err
	}
	if err = r.acquire(c.Local, index); err != nil {
		r.closeEmpty()
		return nil, err
	}
	s := &Session{registry: r, cfg: c, keys: make(map[downKey]bool), leases: map[netip.Addr]bool{c.Local: true}}
	if err = r.state.put("downlinks", c.downKey(), ipv4(c.IPv4), ebpf.UpdateNoExist); err != nil {
		if cleanupErr := s.retire(true); cleanupErr != nil {
			r.orphans = append(r.orphans, s)
			r.queueWarning(cleanupErr)
		}
		r.closeEmpty()
		return nil, err
	}
	s.keys[c.downKey()] = true
	if err = r.state.put("sessions", ipv4(c.IPv4), c.binding(), ebpf.UpdateNoExist); err != nil {
		if cleanupErr := s.retire(true); cleanupErr != nil {
			r.orphans = append(r.orphans, s)
			r.queueWarning(cleanupErr)
		}
		r.closeEmpty()
		return nil, err
	}
	r.sessions[ipv4(c.IPv4)] = s
	r.publishPeers()
	return s, nil
}
func (s *Session) ProgramFD() int {
	s.registry.mu.Lock()
	defer s.registry.mu.Unlock()
	if s.closed || s.registry.collection == nil {
		return -1
	}
	return s.registry.collection.Programs["encap"].FD()
}
func (s *Session) Update(c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	r := s.registry
	r.mu.Lock()
	defer r.unlock()
	if s.closed || s.stopping {
		return errors.New("eBPF session is closed or retiring")
	}
	if c.IPv4 != s.cfg.IPv4 || c.EndpointIfIndex != s.cfg.EndpointIfIndex {
		return errors.New("eBPF handover cannot change the stable UE address or endpoint")
	}
	index, err := r.discover(c)
	if err != nil {
		return err
	}
	if lease := r.locals[c.Local]; lease != nil && lease.ifindex != index {
		return errors.New("N3 interface changed while its binding is active")
	}
	if s.leases[c.Local] {
		lease := r.locals[c.Local]
		if lease == nil || !lease.mapped || !lease.linked {
			return errors.New("target N3 lease is retained for incomplete cleanup and cannot forward")
		}
	}
	newLease := !s.leases[c.Local]
	if newLease {
		if err = r.acquire(c.Local, index); err != nil {
			return err
		}
		s.leases[c.Local] = true
	}
	newKey := !s.keys[c.downKey()]
	rollback := func(cause error) error {
		if newKey && s.keys[c.downKey()] {
			if e := r.state.del("downlinks", c.downKey()); e == nil {
				delete(s.keys, c.downKey())
			} else {
				cause = errors.Join(cause, e)
			}
		}
		if newLease {
			if e := r.release(c.Local); e == nil {
				delete(s.leases, c.Local)
			} else {
				cause = errors.Join(cause, e)
			}
		}
		return cause
	}
	if newKey {
		if err = r.state.put("downlinks", c.downKey(), ipv4(c.IPv4), ebpf.UpdateNoExist); err != nil {
			return rollback(err)
		}
		s.keys[c.downKey()] = true
	}
	if err = r.state.put("sessions", ipv4(c.IPv4), c.binding(), ebpf.UpdateExist); err != nil {
		return rollback(err)
	}
	s.cfg = c // Single canonical replacement commits both uplink and downlink.
	r.publishPeers()
	if err = s.retire(false); err != nil {
		r.queueWarning(fmt.Errorf("eBPF committed handover retains retired resources: %w", err))
	}
	return nil
}
func (s *Session) retire(all bool) error {
	r := s.registry
	var result error
	for k := range s.keys {
		if all || k != s.cfg.downKey() {
			if err := r.state.del("downlinks", k); err != nil {
				result = errors.Join(result, err)
			} else {
				delete(s.keys, k)
			}
		}
	}
	for local := range s.leases {
		if all || local != s.cfg.Local {
			if err := r.release(local); err != nil {
				result = errors.Join(result, err)
			} else {
				delete(s.leases, local)
			}
		}
	}
	return result
}

type peerKey struct{ Local, Remote netip.Addr }

func (r *Registry) publishPeers() {
	peers := make(map[peerKey]bool)
	for _, s := range r.sessions {
		if !s.stopping && !s.closed {
			peers[peerKey{s.cfg.Local, s.cfg.Remote}] = true
		}
	}
	r.peers.Store(&peers)
}
func (r *Registry) reap() {
	for local, port := range r.ports {
		if err := port.Close(); err == nil {
			delete(r.ports, local)
		}
	}
	for index, attachment := range r.links {
		if attachment.refs == 0 {
			if err := attachment.attachment.Close(); err == nil {
				delete(r.links, index)
			}
		}
	}

	pending := r.orphans[:0]
	for _, s := range r.orphans {
		if err := s.retire(true); err != nil {
			pending = append(pending, s)
		}
	}
	r.orphans = pending
}
func (r *Registry) closeEmpty() {
	if len(r.sessions) == 0 && len(r.locals) == 0 && len(r.orphans) == 0 && len(r.ports) == 0 && r.collection != nil {
		for index, attachment := range r.links {
			if attachment.refs == 0 {
				if err := attachment.attachment.Close(); err != nil {
					r.queueWarning(err)
					return
				}
				delete(r.links, index)
			}
		}
		r.collection.Close()
		r.collection = nil
		r.state = nil
	}
}

// Close first removes the canonical mapping. If deactivation fails, the caller
// must retain the endpoint; otherwise a reused interface index could receive an
// old session's packet. An empty tombstone is an additional safe fallback.
func (s *Session) Close() error {
	r := s.registry
	r.mu.Lock()
	defer r.unlock()
	if s.closed {
		return nil
	}
	s.stopping = true
	r.publishPeers()
	key := ipv4(s.cfg.IPv4)
	if err := r.state.del("sessions", key); err != nil {
		tombstoneErr := r.state.put("sessions", key, binding{}, ebpf.UpdateExist)
		return errors.Join(err, tombstoneErr)
	}
	result := s.retire(true)
	// Retirement failures preserve ownership and can be retried with Close.
	if result != nil {
		return result
	}
	delete(r.sessions, key)
	s.closed = true
	r.closeEmpty()
	return nil
}

// Stats reports kernel uplink/downlink successes and owned ingress drops.
// It does not count uplink rejection or packets discarded by later routing.
func (r *Registry) Stats() (uplink, downlink, ingressDrops uint64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.collection == nil {
		return 0, 0, 0, nil
	}
	values := []*uint64{&uplink, &downlink, &ingressDrops}
	for i, p := range values {
		if err = r.collection.Maps["counters"].Lookup(uint32(i), p); err != nil {
			return
		}
	}
	return
}
