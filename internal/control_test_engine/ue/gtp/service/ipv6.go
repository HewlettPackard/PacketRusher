// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	ipv6 "my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"
)

type ipv6NetworkOperations struct {
	addressAdd     func(netlink.Link, *netlink.Addr) error
	addressReplace func(netlink.Link, *netlink.Addr) error
	addressDel     func(netlink.Link, *netlink.Addr) error
	ruleAdd        func(*netlink.Rule) error
	ruleDel        func(*netlink.Rule) error
	routeAdd       func(*netlink.Route) error
	routeDel       func(*netlink.Route) error
}

var ipv6Network = ipv6NetworkOperations{
	addressAdd: netlink.AddrAdd, addressReplace: netlink.AddrReplace, addressDel: netlink.AddrDel,
	ruleAdd: netlink.RuleAdd, ruleDel: netlink.RuleDel, routeAdd: netlink.RouteReplace, routeDel: netlink.RouteDel,
}

type ipv6Binding struct {
	link      netlink.Link
	table     uint32
	vrf       bool
	session   *context.UEPDUSession
	ops       ipv6NetworkOperations
	addresses []*netlink.Addr
	rules     []*netlink.Rule
	route     *netlink.Route
	address   *netlink.Addr
	rule      *netlink.Rule
}

// install stages a new prefix before replacing its route; failed staging leaves
// the previous prefix routed. Addresses and policies remain tracked even if a
// cleanup call fails, so final cleanup can retry and quarantine the table.
func (b *ipv6Binding) install(advertisement ipv6.Advertisement) error {
	address := &netlink.Addr{IPNet: addressNet(advertisement.Address), Flags: unix.IFA_F_NODAD, PreferedLft: lifetime(advertisement.PreferredLifetime), ValidLft: lifetime(advertisement.ValidLifetime)}
	if b.address != nil && b.address.IP.Equal(address.IP) {
		if err := b.ops.addressReplace(b.link, address); err != nil {
			return fmt.Errorf("renew UE IPv6 prefix: %w", err)
		}
		b.address = address
		return nil
	}
	if err := b.ops.addressAdd(b.link, address); err != nil {
		return fmt.Errorf("install UE IPv6 address: %w", err)
	}
	b.addresses = append(b.addresses, address)
	var rule *netlink.Rule
	if !b.vrf {
		rule = netlink.NewRule()
		rule.Family = netlink.FAMILY_V6
		rule.Priority = 100
		rule.Table = int(b.table)
		rule.Src = address.IPNet
		if err := b.ops.ruleAdd(rule); err != nil {
			return errors.Join(fmt.Errorf("install IPv6 source routing policy: %w", err), b.removeAddress(address))
		}
		b.rules = append(b.rules, rule)
	}
	route := &netlink.Route{Family: netlink.FAMILY_V6, Dst: &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}, LinkIndex: b.link.Attrs().Index, Table: int(b.table), Scope: netlink.SCOPE_LINK, Protocol: 4, Priority: 1, Src: net.IP(advertisement.Address.AsSlice())}
	if err := b.ops.routeAdd(route); err != nil {
		return errors.Join(fmt.Errorf("install IPv6 default route: %w", err), b.removeRule(rule), b.removeAddress(address))
	}
	oldRule, oldAddress := b.rule, b.address
	b.route, b.rule, b.address = route, rule, address
	if err := b.session.SetIPv6(advertisement.Address); err != nil {
		return err
	}
	return errors.Join(b.removeRule(oldRule), b.removeAddress(oldAddress))
}
func (b *ipv6Binding) removeRule(rule *netlink.Rule) error {
	if rule == nil {
		return nil
	}
	if err := b.ops.ruleDel(rule); !routingObjectRemoved(err) {
		return err
	}
	for i, candidate := range b.rules {
		if candidate == rule {
			b.rules = append(b.rules[:i], b.rules[i+1:]...)
			break
		}
	}
	return nil
}
func (b *ipv6Binding) removeAddress(address *netlink.Addr) error {
	if address == nil {
		return nil
	}
	if err := b.ops.addressDel(b.link, address); !routingObjectRemoved(err) {
		return err
	}
	for i, candidate := range b.addresses {
		if candidate == address || candidate.IP.Equal(address.IP) {
			b.addresses = append(b.addresses[:i], b.addresses[i+1:]...)
			break
		}
	}
	return nil
}
func (b *ipv6Binding) withdraw() error {
	var failures []error
	if b.route != nil {
		if err := b.ops.routeDel(b.route); !routingObjectRemoved(err) {
			failures = append(failures, err)
		} else {
			b.route = nil
		}
	}
	if err := b.removeRule(b.rule); err != nil {
		failures = append(failures, err)
	} else {
		b.rule = nil
	}
	if err := b.removeAddress(b.address); err != nil {
		failures = append(failures, err)
	} else {
		b.address = nil
	}
	b.session.ClearIPv6()
	return errors.Join(failures...)
}
func (b *ipv6Binding) cleanup() error {
	var failures []error
	if err := b.withdraw(); err != nil {
		failures = append(failures, err)
	}
	for _, rule := range append([]*netlink.Rule(nil), b.rules...) {
		if err := b.removeRule(rule); err != nil {
			failures = append(failures, err)
		}
	}
	for _, address := range append([]*netlink.Addr(nil), b.addresses...) {
		if err := b.removeAddress(address); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// SetupIPv6Session discovers the UPF's prefix, then refreshes advertisements and
// withdraws expired routing. Cleanup must precede userspace transport/device/
// table cleanup. Even on failure a non-nil cleanup must be retained and called.
func SetupIPv6Session(ue *context.UEContext, session *context.UEPDUSession, link netlink.Link, table uint32, send func([]byte) error, advertisements <-chan []byte) (func() error, error) {
	return setupIPv6Session(ue, session, link, table, send, advertisements, ipv6Network, 4*time.Second)
}
func setupIPv6Session(ue *context.UEContext, session *context.UEPDUSession, link netlink.Link, table uint32, send func([]byte) error, advertisements <-chan []byte, ops ipv6NetworkOperations, retryDelay time.Duration) (cleanup func() error, resultError error) {
	iid, enabled := session.GetIPv6InterfaceID()
	if !enabled {
		return func() error { return nil }, nil
	}
	if link == nil || link.Attrs() == nil || link.Attrs().MTU < 1280 || table < firstRoutingTable || send == nil || advertisements == nil {
		return nil, fmt.Errorf("IPv6 requires a session transport, routing table, and TUN MTU of at least 1280")
	}
	binding := &ipv6Binding{link: link, table: table, vrf: ue.TunnelMode == config.TunnelVrf, session: session, ops: ops}
	stop, done := make(chan struct{}), make(chan struct{})
	var started bool
	var once sync.Once
	var cleanupLock sync.Mutex
	cleanup = func() error {
		cleanupLock.Lock()
		defer cleanupLock.Unlock()
		once.Do(func() {
			close(stop)
			if started {
				<-done
			}
		})
		return binding.cleanup()
	}
	defer func() {
		if resultError != nil {
			resultError = errors.Join(resultError, cleanup())
		}
	}()
	linkLocal := &netlink.Addr{IPNet: addressNet(ipv6.LinkLocal(iid)), Flags: unix.IFA_F_NODAD}
	if err := ops.addressAdd(link, linkLocal); err != nil {
		return cleanup, fmt.Errorf("install allocated IPv6 link-local address: %w", err)
	}
	binding.addresses = append(binding.addresses, linkLocal)
	var advertisement ipv6.Advertisement
	for attempt := 0; attempt < 3 && !advertisement.Address.IsValid(); attempt++ {
		if err := send(ipv6.RouterSolicitation(iid)); err != nil {
			return cleanup, fmt.Errorf("send IPv6 Router Solicitation: %w", err)
		}
		timer := time.NewTimer(retryDelay)
		waiting := true
		for waiting {
			select {
			case packet, open := <-advertisements:
				if !open {
					timer.Stop()
					return cleanup, fmt.Errorf("IPv6 transport ended before prefix discovery")
				}
				if candidate, err := ipv6.ParseAdvertisement(packet, iid); err == nil && candidate.ValidLifetime > 0 && candidate.RouterLifetime > 0 {
					advertisement = candidate
					waiting = false
				}
			case <-timer.C:
				waiting = false
			case <-ue.Done():
				timer.Stop()
				return cleanup, fmt.Errorf("UE stopped during IPv6 prefix discovery")
			}
		}
		timer.Stop()
	}
	if !advertisement.Address.IsValid() {
		return cleanup, fmt.Errorf("UPF did not advertise a usable IPv6 /64 prefix after three solicitations")
	}
	if err := binding.install(advertisement); err != nil {
		return cleanup, err
	}
	started = true
	go func() {
		defer close(done)
		expiry := time.NewTimer(advertisementLifetime(advertisement))
		defer expiry.Stop()
		refresh := time.NewTimer(advertisementLifetime(advertisement) / 2)
		defer refresh.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ue.Done():
				return
			case packet, open := <-advertisements:
				if !open {
					return
				}
				candidate, err := ipv6.ParseAdvertisement(packet, iid)
				if err != nil {
					continue
				}
				if candidate.RouterLifetime == 0 || candidate.ValidLifetime == 0 {
					allocation := session.GetIPv6()
					if candidate.RouterLifetime == 0 || (allocation.IsValid() && candidate.Prefix.Contains(allocation)) {
						if err := binding.withdraw(); err != nil {
							log.Warn("[UE][GTP] Withdrawn IPv6 prefix cleanup failed: ", err)
						}
						expiry.Stop()
						refresh.Reset(retryDelay)
					}
					continue
				}
				if err = binding.install(candidate); err != nil {
					log.Warn("[UE][GTP] IPv6 advertisement update failed: ", err)
					continue
				}
				expiry.Reset(advertisementLifetime(candidate))
				refresh.Reset(advertisementLifetime(candidate) / 2)
			case <-refresh.C:
				if err := send(ipv6.RouterSolicitation(iid)); err != nil {
					log.Warn("[UE][GTP] IPv6 prefix refresh solicitation failed: ", err)
				}
				refresh.Reset(retryDelay)
			case <-expiry.C:
				if err := binding.withdraw(); err != nil {
					log.Warn("[UE][GTP] Expired IPv6 prefix cleanup failed: ", err)
				}
			}
		}
	}()
	return cleanup, nil
}
func advertisementLifetime(a ipv6.Advertisement) time.Duration {
	lifetime := a.ValidLifetime
	if uint32(a.RouterLifetime) < lifetime {
		lifetime = uint32(a.RouterLifetime)
	}
	return time.Duration(lifetime) * time.Second
}
func addressNet(address netip.Addr) *net.IPNet {
	return &net.IPNet{IP: net.IP(address.AsSlice()), Mask: net.CIDRMask(128, 128)}
}
func lifetime(seconds uint32) int {
	if uint64(seconds) > uint64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	return int(seconds)
}
