//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
)

//go:embed gtpu_bpfel.o
var object []byte

// ConfigureEndpoint disables checksum offload and limits TCP software GSO before
// the owned TCX egress hook. A GTP peer receives bytes, not inner skb metadata.
// Each GTP-U length describes one complete inner IPv4 or IPv6 packet.
// Call only for the backend's newly owned TUN, before installing its route.
func ConfigureEndpoint(endpoint netlink.Link, port userspace.PacketPort) error {
	owned, ok := port.(interface {
		SyscallConn() (syscall.RawConn, error)
	})
	if !ok {
		return errors.New("eBPF endpoint has no owned TUN descriptor")
	}
	raw, err := owned.SyscallConn()
	if err != nil {
		return err
	}
	var ioctlErr error
	err = raw.Control(func(fd uintptr) { ioctlErr = unix.IoctlSetInt(int(fd), unix.TUNSETOFFLOAD, 0) })
	if err = errors.Join(err, ioctlErr); err != nil {
		return fmt.Errorf("disable owned eBPF TUN checksum offload: %w", err)
	}
	if err := netlink.LinkSetGSOMaxSegs(endpoint, 1); err != nil {
		return fmt.Errorf("limit eBPF endpoint TCP segmentation: %w", err)
	}
	return nil
}

// Config binds a dual-stack PDU session to an owned, stable L3 endpoint.
type Config struct {
	Local, Remote, IPv4      netip.Addr
	AllowIPv6                bool
	IPv6InterfaceID          [8]byte
	IPv6                     netip.Addr
	Inject                   func([]byte) error
	UplinkTEID, DownlinkTEID uint32
	QFI                      uint8
	EndpointIfIndex, MTU     int
	stageTX                  uint32
}

func (c Config) Validate() error {
	for _, a := range []netip.Addr{c.Local, c.Remote} {
		if !a.Is4() || a.IsUnspecified() || a.IsMulticast() {
			return errors.New("eBPF requires unicast IPv4 UE and N3 addresses")
		}
	}
	if c.Local == c.Remote || c.IPv4 == c.Local || c.IPv4 == c.Remote {
		return errors.New("eBPF UE, local N3 and peer addresses must be distinct")
	}
	if c.IPv4.IsValid() && (!c.IPv4.Is4() || c.IPv4.IsUnspecified() || c.IPv4.IsMulticast() || c.IPv4.IsLoopback()) {
		return errors.New("invalid UE IPv4 address")
	}
	if !c.IPv4.IsValid() && !c.AllowIPv6 {
		return errors.New("session has neither IPv4 nor IPv6")
	}
	if c.IPv6.IsValid() && (!c.AllowIPv6 || !c.IPv6.Is6() || !c.IPv6.IsGlobalUnicast()) {
		return errors.New("invalid IPv6 allocation")
	}
	if c.IPv6.IsValid() {
		a := c.IPv6.As16()
		if !bytes.Equal(a[8:], c.IPv6InterfaceID[:]) {
			return errors.New("IPv6 allocation must use the NAS IID")
		}
	}
	if c.AllowIPv6 && c.MTU < 1280 {
		return errors.New("IPv6 endpoint MTU must be at least 1280")
	}
	if c.UplinkTEID == 0 || c.DownlinkTEID == 0 || c.QFI > 63 || c.EndpointIfIndex <= 0 || c.MTU < 68 || c.MTU > 65491 {
		return errors.New("invalid eBPF TEID, QFI, endpoint or IPv4 N3 payload MTU (68..65491)")
	}
	return nil
}

type binding struct {
	Local, Peer, UE, UplinkTEID, DownlinkTEID, Endpoint, MTU, NextHop, StageTX uint32
	QFI, IPv6, PrefixValid, Padding                                            uint8
	IID, Prefix                                                                [8]byte
}
type downKey struct{ Local, Peer, TEID uint32 }

func ipv4(a netip.Addr) uint32    { b := a.As4(); return binary.LittleEndian.Uint32(b[:]) }
func (c Config) identity() uint32 { return uint32(c.EndpointIfIndex) }

type n3Path struct {
	IfIndex  int
	NextHop  netip.Addr
	Loopback bool
}

func (c Config) binding(paths ...n3Path) binding {
	b := binding{Local: ipv4(c.Local), Peer: ipv4(c.Remote), UplinkTEID: c.UplinkTEID, DownlinkTEID: c.DownlinkTEID, Endpoint: c.identity(), MTU: uint32(c.MTU), QFI: c.QFI, IID: c.IPv6InterfaceID, NextHop: ipv4(c.Remote), StageTX: c.stageTX}
	if len(paths) > 0 && paths[0].NextHop.IsValid() {
		b.NextHop = ipv4(paths[0].NextHop)
	}
	if c.IPv4.Is4() {
		b.UE = ipv4(c.IPv4)
	}
	if c.AllowIPv6 {
		b.IPv6 = 1
	}
	if c.IPv6.IsValid() {
		address := c.IPv6.As16()
		copy(b.Prefix[:], address[:8])
		b.PrefixValid = 1
	}
	return b
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
// Its socket worker answers Echo and handles Router Advertisements plus the
// kernel-validated checksum/reassembly/optional-header fallback; ordinary packets use TCX.
type Registry struct {
	peers           atomic.Pointer[map[peerKey]bool]
	controls        atomic.Pointer[map[downKey]controlBinding]
	mu              sync.Mutex
	collection      *ebpf.Collection
	state           store
	sessions        map[uint32]*Session
	orphans         []*Session
	ports           map[netip.Addr]closer
	locals          map[netip.Addr]*n3Lease
	links           map[int]*ingress
	discover        func(Config) (n3Path, error)
	listen          func(netip.Addr) (closer, error)
	attach          func(int) (closer, error)
	attachEndpoint  func(int) (closer, error)
	warning         func(error)
	pendingWarnings []error
}

func NewRegistry() *Registry {
	r := &Registry{ports: make(map[netip.Addr]closer), sessions: make(map[uint32]*Session), locals: make(map[netip.Addr]*n3Lease), links: make(map[int]*ingress), discover: discoverN3, warning: func(error) {}}
	r.listen = func(a netip.Addr) (closer, error) {
		return newManagementSocket(a, func(peer netip.AddrPort) bool {
			snapshot := r.peers.Load()
			return peer.Port() == 2152 && snapshot != nil && (*snapshot)[peerKey{a, peer.Addr()}]
		}, func(peer netip.AddrPort, packet []byte) { r.receiveControl(a, peer, packet) })
	}
	r.attach = func(index int) (closer, error) {
		return link.AttachTCX(link.TCXOptions{Interface: index, Program: r.collection.Programs["decap"], Attach: ebpf.AttachTCXIngress, Anchor: link.Tail()})
	}
	r.attachEndpoint = func(index int) (closer, error) {
		stage, err := attachChecksumStage(index, r.collection.Programs["encap"], r.collection.Programs["relay"])
		if stage == nil {
			return nil, err
		}
		return stage, err
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
	spec.Programs["encap"].AttachType = ebpf.AttachTCXEgress
	spec.Programs["relay"].AttachType = ebpf.AttachTCXIngress
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("load PacketRusher eBPF programs (CAP_BPF/CAP_NET_ADMIN and TCX support required): %w", err)
	}
	r.collection = collection
	r.state = kernelStore{collection.Maps}
	return nil
}
func discoverN3(c Config) (n3Path, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return n3Path{}, err
	}
	path := n3Path{}
	for _, l := range links {
		addresses, err := netlink.AddrList(l, netlink.FAMILY_V4)
		if err != nil {
			return n3Path{}, err
		}
		for _, a := range addresses {
			if !a.IP.Equal(net.IP(c.Local.AsSlice())) {
				continue
			}
			if path.IfIndex != 0 {
				return n3Path{}, errors.New("N3 address is assigned to multiple interfaces")
			}
			at := l.Attrs()
			loopback := at.Flags&net.FlagLoopback != 0
			if !loopback && (at.EncapType != "ether" || len(at.HardwareAddr) != 6 || at.MasterIndex != 0) {
				return n3Path{}, errors.New("N3 requires a loopback or Ethernet interface without a master")
			}
			if at.Flags&net.FlagUp == 0 || c.MTU+44 > at.MTU {
				return n3Path{}, errors.New("N3 interface is down or lacks negotiated tunnel MTU headroom")
			}
			path.IfIndex, path.Loopback = at.Index, loopback
		}
	}
	if path.IfIndex == 0 {
		return n3Path{}, fmt.Errorf("N3 address %s is not assigned", c.Local)
	}
	routes, err := netlink.RouteGetWithOptions(net.IP(c.Remote.AsSlice()), &netlink.RouteGetOptions{SrcAddr: net.IP(c.Local.AsSlice())})
	if err != nil {
		return n3Path{}, err
	}
	if len(routes) != 1 || routes[0].LinkIndex != path.IfIndex || len(routes[0].MultiPath) != 0 {
		return n3Path{}, errors.New("N3 requires one route through its assigned interface")
	}
	if routes[0].MTU > 0 && c.MTU+44 > routes[0].MTU {
		return n3Path{}, errors.New("N3 route MTU is smaller than the negotiated tunnel")
	}
	path.NextHop = c.Remote
	if len(routes[0].Gw) > 0 {
		addr, ok := netip.AddrFromSlice(routes[0].Gw)
		if !ok || !addr.Unmap().Is4() {
			return n3Path{}, errors.New("N3 next hop must be IPv4")
		}
		path.NextHop = addr.Unmap()
	}
	return path, nil
}
func (r *Registry) acquire(local netip.Addr, path n3Path) error {
	index := path.IfIndex
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
	if err = r.state.put("locals", ipv4(local), ownedN3Index(path), ebpf.UpdateNoExist); err != nil {
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

// Session keeps the same endpoint and owned checksum stage across handover.
type Session struct {
	registry       *Registry
	path           n3Path
	cfg            Config
	keys           map[downKey]bool
	leases         map[netip.Addr]bool
	closed         bool
	stopping       bool
	endpoint       closer
	stageKey       uint32
	stageMapped    bool
	advertisements chan []byte
}

func (r *Registry) Open(c Config) (*Session, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.unlock()
	if r.sessions[c.identity()] != nil {
		return nil, errors.New("eBPF endpoint already has a session")
	}
	r.reap()
	path, err := r.discover(c)
	if err != nil {
		return nil, err
	}
	if err = r.load(); err != nil {
		return nil, err
	}
	if err = r.acquire(c.Local, path); err != nil {
		r.closeEmpty()
		return nil, err
	}
	s := &Session{registry: r, cfg: c, path: path, advertisements: make(chan []byte, 8), keys: make(map[downKey]bool), leases: map[netip.Addr]bool{c.Local: true}}
	if err = r.state.put("downlinks", c.downKey(), c.identity(), ebpf.UpdateNoExist); err != nil {
		if cleanupErr := s.retire(true); cleanupErr != nil {
			r.orphans = append(r.orphans, s)
			r.queueWarning(cleanupErr)
		}
		r.closeEmpty()
		return nil, err
	}
	s.keys[c.downKey()] = true
	s.endpoint, err = r.attachEndpoint(c.EndpointIfIndex)
	if err != nil {
		if cleanupErr := s.retire(true); cleanupErr != nil {
			r.orphans = append(r.orphans, s)
			r.queueWarning(cleanupErr)
		}
		r.closeEmpty()
		return nil, fmt.Errorf("attach eBPF UE egress: %w", err)
	}
	if staged, ok := s.endpoint.(interface {
		TransmitIndex() int
		PeerIndex() int
	}); ok {
		c.stageTX = uint32(staged.TransmitIndex())
		s.cfg = c
		s.stageKey = uint32(staged.PeerIndex())
		if err = r.state.put("stages", s.stageKey, c.identity(), ebpf.UpdateNoExist); err != nil {
			if cleanupErr := s.retire(true); cleanupErr != nil {
				r.orphans = append(r.orphans, s)
				r.queueWarning(cleanupErr)
			}
			r.closeEmpty()
			return nil, err
		}
		s.stageMapped = true
	}
	if err = r.state.put("sessions", c.identity(), c.binding(path), ebpf.UpdateNoExist); err != nil {
		if cleanupErr := s.retire(true); cleanupErr != nil {
			r.orphans = append(r.orphans, s)
			r.queueWarning(cleanupErr)
		}
		r.closeEmpty()
		return nil, err
	}
	r.sessions[c.identity()] = s
	r.publishPeers()
	return s, nil
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
	// Prefix changes are owned by the allocation callback, not a racing handover snapshot.
	c.IPv6 = s.cfg.IPv6
	c.stageTX = s.cfg.stageTX
	if c.IPv4 != s.cfg.IPv4 || c.EndpointIfIndex != s.cfg.EndpointIfIndex || c.AllowIPv6 != s.cfg.AllowIPv6 || c.IPv6InterfaceID != s.cfg.IPv6InterfaceID {
		return errors.New("eBPF handover cannot change the stable UE address or endpoint")
	}
	path, err := r.discover(c)
	index := path.IfIndex
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
		if err = r.acquire(c.Local, path); err != nil {
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
		if err = r.state.put("downlinks", c.downKey(), c.identity(), ebpf.UpdateNoExist); err != nil {
			return rollback(err)
		}
		s.keys[c.downKey()] = true
	}
	if err = r.state.put("sessions", c.identity(), c.binding(path), ebpf.UpdateExist); err != nil {
		return rollback(err)
	}
	s.cfg, s.path = c, path // Single canonical replacement commits both uplink and downlink.
	r.publishPeers()
	if err = s.retire(false); err != nil {
		r.queueWarning(fmt.Errorf("eBPF committed handover retains retired resources: %w", err))
	}
	return nil
}
func (s *Session) retire(all bool) error {
	r := s.registry
	var result error
	if all && s.stageMapped {
		if err := r.state.del("stages", s.stageKey); err != nil {
			return err
		}
		s.stageMapped = false
	}
	if all && s.endpoint != nil {
		if err := s.endpoint.Close(); err != nil {
			return err
		}
		s.endpoint = nil
	}
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
	controls := make(map[downKey]controlBinding)
	for _, s := range r.sessions {
		if !s.stopping && !s.closed {
			peers[peerKey{s.cfg.Local, s.cfg.Remote}] = true
			controls[s.cfg.downKey()] = controlBinding{s, s.cfg}
		}
	}
	r.peers.Store(&peers)
	r.controls.Store(&controls)
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
	key := s.cfg.identity()
	if err := r.state.del("sessions", key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
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

// Stats reports kernel uplink/downlink redirect attempts and owned ingress drops.
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
