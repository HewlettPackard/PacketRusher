//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"net/netip"

	"github.com/cilium/ebpf"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

func ownedN3Index(path n3Path) uint32 {
	index := uint32(path.IfIndex)
	if path.Loopback {
		index |= 1 << 31
	}
	return index
}

// These immutable snapshots keep the joined socket worker independent of the
// registry lock. Final lease release can join it while holding that lock.
type controlBinding struct {
	session *Session
	config  Config
}

func (c Config) allowsIPv6(address netip.Addr) bool {
	if !c.AllowIPv6 || !address.Is6() {
		return false
	}
	a := address.As16()
	if !bytes.Equal(a[8:], c.IPv6InterfaceID[:]) {
		return false
	}
	if address == ipv6.LinkLocal(c.IPv6InterfaceID) {
		return true
	}
	return address.IsGlobalUnicast() && c.IPv6.IsValid() && netip.PrefixFrom(c.IPv6, 64).Contains(address)
}

func (r *Registry) receiveControl(local netip.Addr, peer netip.AddrPort, packet []byte) {
	if peer.Port() != 2152 {
		return
	}
	teid, payload, err := userspace.Decode(packet)
	if err != nil {
		return
	}
	snapshot := r.controls.Load()
	if snapshot == nil {
		return
	}
	held, ok := (*snapshot)[downKey{ipv4(local), ipv4(peer.Addr()), teid}]
	if !ok || len(payload) > held.config.MTU {
		return
	}
	// Reassembled datagrams still require the same strict optional container
	// and QFI as kernel-forwarded traffic. Never bypass that boundary in Go.
	switch packet[0] {
	case 0x30:
	case 0x32:
		if len(packet) < 12 || packet[10] != 0 || packet[11] != 0 {
			return
		}
	case 0x34, 0x36:
		if len(packet) < 16 || (packet[0] == 0x34 && binary.BigEndian.Uint16(packet[8:10]) != 0) || packet[10] != 0 || packet[11] != 0x85 || packet[12] != 1 || packet[13] != 0 || packet[14] != held.config.QFI || packet[15] != 0 {
			return
		}
	default:
		return
	}
	if ipv6.RouterAdvertisement(payload) {
		if _, err := ipv6.ParseAdvertisement(payload, held.config.IPv6InterfaceID); (err != nil && !errors.Is(err, ipv6.ErrNoAutonomousPrefix)) || !held.config.AllowIPv6 {
			return
		}
		select {
		case held.session.advertisements <- append([]byte(nil), payload...):
		default:
		}
		return
	}
	_, destination, err := userspace.IPAddresses(payload)
	if err != nil {
		return
	}
	address, ok := netip.AddrFromSlice(destination)
	if !ok || (address.Is4() && address != held.config.IPv4) || (address.Is6() && !held.config.allowsIPv6(address)) {
		return
	}
	// Outer fragments were reassembled by the kernel UDP stack. The owner
	// supplies a nonblocking TUN injector; ordinary traffic stays in TCX.
	if held.config.Inject != nil {
		if err := held.config.Inject(payload); err != nil {
			log.Warn("[UE][eBPF] Reassembled downlink injection failed: ", err)
		}
	}
}

func (s *Session) Advertisements() <-chan []byte { return s.advertisements }

func (s *Session) SendControl(payload []byte) error {
	r := s.registry
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.closed || s.stopping {
		return net.ErrClosed
	}
	if len(payload) < 48 || payload[0]>>4 != 6 || payload[6] != 58 || payload[40] != 133 {
		return errors.New("only allocated IPv6 Router Solicitations use the control sender")
	}
	source, _, err := userspace.IPAddresses(payload)
	if err != nil {
		return err
	}
	address, ok := netip.AddrFromSlice(source)
	if !ok || !s.cfg.allowsIPv6(address) {
		return errors.New("solicitation source does not belong to the allocated IID")
	}
	packet, err := userspace.Encode(s.cfg.UplinkTEID, s.cfg.QFI, payload)
	if err != nil {
		return err
	}
	lease := r.locals[s.cfg.Local]
	if lease == nil {
		return net.ErrClosed
	}
	writer, ok := lease.socket.(interface {
		Send([]byte, netip.AddrPort) error
	})
	if !ok {
		return errors.New("management transport does not support control sends")
	}
	return writer.Send(packet, netip.AddrPortFrom(s.cfg.Remote, 2152))
}

// SetIPv6 is the sole prefix authorization update. A failed map write retires
// canonical forwarding; the service additionally disables/quarantines its TUN
// if kernel deactivation itself fails.
func (s *Session) SetIPv6(address netip.Addr) error {
	r := s.registry
	r.mu.Lock()
	defer r.unlock()
	if s.closed {
		if !address.IsValid() {
			return nil
		}
		return net.ErrClosed
	}
	if s.stopping {
		if address.IsValid() {
			return net.ErrClosed
		}
		// Retrying revocation may retire a previously failed canonical delete.
		err := r.state.del("sessions", s.cfg.identity())
		if errors.Is(err, ebpf.ErrKeyNotExist) {
			err = nil
		}
		if err != nil {
			err = errors.Join(err, r.state.put("sessions", s.cfg.identity(), binding{}, ebpf.UpdateExist))
		}
		if err == nil {
			s.cfg.IPv6 = netip.Addr{}
		}
		return err
	}
	next := s.cfg
	next.IPv6 = address
	if err := next.Validate(); err != nil {
		return err
	}
	if err := r.state.put("sessions", next.identity(), next.binding(s.path), ebpf.UpdateExist); err != nil {
		s.stopping = true
		r.publishPeers()
		deactivate := r.state.del("sessions", next.identity())
		if deactivate != nil {
			deactivate = errors.Join(deactivate, r.state.put("sessions", next.identity(), binding{}, ebpf.UpdateExist))
		}
		return errors.Join(err, deactivate)
	}
	s.cfg = next
	r.publishPeers()
	return nil
}

// SendUplink accepts only oversized packets completed by the owned TUN driver.
// The actor's current mapping decides peer/TEIDs at dispatch, including handover
// or prefix revocation while the descriptor had a queued packet.
func (s *Session) SendUplink(payload []byte) error {
	r := s.registry
	r.mu.Lock()
	defer r.unlock()
	if s.closed || s.stopping {
		return net.ErrClosed
	}
	if len(payload) <= 8000 || len(payload) > s.cfg.MTU {
		return errors.New("invalid oversized eBPF fallback payload")
	}
	source, _, err := userspace.IPAddresses(payload)
	if err != nil {
		return err
	}
	address, ok := netip.AddrFromSlice(source)
	if !ok || (address.Is4() && address != s.cfg.IPv4) || (address.Is6() && !s.cfg.allowsIPv6(address)) {
		return errors.New("fallback source does not belong to current allocation")
	}
	packet, err := userspace.Encode(s.cfg.UplinkTEID, s.cfg.QFI, payload)
	if err != nil {
		return err
	}
	lease := r.locals[s.cfg.Local]
	if lease == nil {
		return net.ErrClosed
	}
	writer, ok := lease.socket.(interface {
		Send([]byte, netip.AddrPort) error
	})
	if !ok {
		return errors.New("owned N3 transport cannot send jumbo payload")
	}
	return writer.Send(packet, netip.AddrPortFrom(s.cfg.Remote, 2152))
}
