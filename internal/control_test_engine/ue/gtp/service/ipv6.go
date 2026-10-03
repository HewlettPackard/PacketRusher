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
	link                netlink.Link
	table               uint32
	vrf                 bool
	session             *context.UEPDUSession
	ops                 ipv6NetworkOperations
	addresses           []*netlink.Addr
	rules               []*netlink.Rule
	route               *netlink.Route
	claim               *netlink.Route
	address             *netlink.Addr
	rule                *netlink.Rule
	allocationCallbacks []func(netip.Addr) error
}

// A forwarding owner must either commit the allocation or disable its path on
// error. Notifications run synchronously in the discovery worker, without any
// transport locks. Invalid addresses revoke an allocation, even when network
// object retirement fails and its source policy must remain installed.
func (b *ipv6Binding) notifyAllocation(address netip.Addr) error {
	var failures []error
	for _, callback := range b.allocationCallbacks {
		if callback != nil {
			if err := callback(address); err != nil {
				failures = append(failures, fmt.Errorf("update IPv6 forwarding allocation: %w", err))
			}
		}
	}
	return errors.Join(failures...)
}

// install stages a new prefix before replacing its route; failed staging leaves
// the previous prefix routed. Addresses and policies remain tracked even if a
// cleanup call fails, so final cleanup can retry and quarantine the table.
func (b *ipv6Binding) install(advertisement ipv6.Advertisement) (bool, error) {
	address := &netlink.Addr{IPNet: addressNet(advertisement.Address), Flags: unix.IFA_F_NODAD, PreferedLft: lifetime(advertisement.PreferredLifetime), ValidLft: lifetime(advertisement.ValidLifetime)}
	sameAddress := b.address != nil && b.address.IP.Equal(address.IP)
	rule := b.rule
	if !sameAddress && !b.vrf {
		rule = netlink.NewRule()
		rule.Family = netlink.FAMILY_V6
		rule.Priority = 100
		rule.Table = int(b.table)
		rule.Src = address.IPNet
		if err := b.ops.ruleAdd(rule); err != nil {
			return false, fmt.Errorf("install IPv6 source routing policy: %w", err)
		}
		b.rules = append(b.rules, rule)
	}
	if sameAddress {
		if err := b.ops.addressReplace(b.link, address); err != nil {
			return false, fmt.Errorf("renew UE IPv6 prefix: %w", err)
		}
	} else {
		if err := b.ops.addressAdd(b.link, address); err != nil {
			return false, errors.Join(fmt.Errorf("install UE IPv6 address: %w", err), b.removeRule(rule))
		}
		b.addresses = append(b.addresses, address)
	}
	if err := b.updateRouter(advertisement.Address, advertisement.RouterLifetime); err != nil {
		if sameAddress {
			return true, err
		}
		return false, errors.Join(err, b.retire(address, rule))
	}
	if err := b.notifyAllocation(advertisement.Address); err != nil {
		// Staging is not publication. Revoke any partially notified forwarding
		// owner and remove the candidate, retaining failed deletions for cleanup.
		withdrawErr := b.withdraw()
		if !sameAddress {
			withdrawErr = errors.Join(withdrawErr, b.retire(address, rule))
		}
		return false, errors.Join(err, withdrawErr)
	}
	oldRule, oldAddress := b.rule, b.address
	b.rule, b.address = rule, address
	if err := b.session.SetIPv6(advertisement.Address); err != nil {
		if len(b.allocationCallbacks) > 0 {
			return false, errors.Join(err, b.withdraw(), b.retire(oldAddress, oldRule))
		}
		return true, err
	}
	if sameAddress {
		return true, nil
	}
	return true, b.retire(oldAddress, oldRule)
}

// Keep an address's source policy until the address has gone. A failed deletion
// must not leave its source eligible for the host's main routing table.
func (b *ipv6Binding) retire(address *netlink.Addr, rule *netlink.Rule) error {
	if err := b.removeAddress(address); err != nil {
		return err
	}
	return b.removeRule(rule)
}

func (b *ipv6Binding) updateRouter(address netip.Addr, lifetime uint16) error {
	if lifetime == 0 {
		return b.withdrawRouter()
	}
	route := &netlink.Route{Family: netlink.FAMILY_V6, Dst: &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}, LinkIndex: b.link.Attrs().Index, Table: int(b.table), Scope: netlink.SCOPE_LINK, Protocol: 4, Priority: 1, Src: net.IP(address.AsSlice())}
	if err := b.ops.routeAdd(route); err != nil {
		return fmt.Errorf("install IPv6 default router: %w", err)
	}
	b.route = route
	return nil
}
func (b *ipv6Binding) withdrawRouter() error {
	if b.route == nil {
		return nil
	}
	if err := b.ops.routeDel(b.route); !routingObjectRemoved(err) {
		return err
	}
	b.route = nil
	return nil
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
	if err := b.notifyAllocation(netip.Addr{}); err != nil {
		failures = append(failures, err)
	}
	if err := b.withdrawRouter(); err != nil {
		failures = append(failures, err)
	}
	if err := b.removeAddress(b.address); err != nil {
		failures = append(failures, err)
	} else {
		b.address = nil
		if err := b.removeRule(b.rule); err != nil {
			failures = append(failures, err)
		} else {
			b.rule = nil
		}
	}
	b.session.ClearIPv6()
	return errors.Join(failures...)
}
func (b *ipv6Binding) cleanup() error {
	var failures []error
	if err := b.withdraw(); err != nil {
		failures = append(failures, err)
	}
	for _, address := range append([]*netlink.Addr(nil), b.addresses...) {
		if err := b.removeAddress(address); err != nil {
			failures = append(failures, err)
		}
	}
	for _, rule := range append([]*netlink.Rule(nil), b.rules...) {
		addressPresent := false
		for _, address := range b.addresses {
			if rule.Src != nil && rule.Src.IP.Equal(address.IP) {
				addressPresent = true
				break
			}
		}
		if !addressPresent {
			if err := b.removeRule(rule); err != nil {
				failures = append(failures, err)
			}
		}
	}
	if len(failures) == 0 && b.claim != nil {
		if err := b.ops.routeDel(b.claim); !routingObjectRemoved(err) {
			failures = append(failures, err)
		} else {
			b.claim = nil
		}
	}
	return errors.Join(failures...)
}

// SetupIPv6Session discovers the UPF's prefix, then refreshes advertisements and
// withdraws expired routing. Optional forwarding callbacks must be bounded and
// disable forwarding when their update fails. Cleanup must precede transport,
// device and table cleanup. Retain and call non-nil cleanup even on failure.
func SetupIPv6Session(ue *context.UEContext, session *context.UEPDUSession, link netlink.Link, table uint32, send func([]byte) error, advertisements <-chan []byte, callbacks ...func(netip.Addr) error) (func() error, error) {
	return setupIPv6Session(ue, session, link, table, send, advertisements, ipv6Network, 4*time.Second, callbacks...)
}
func setupIPv6Session(ue *context.UEContext, session *context.UEPDUSession, link netlink.Link, table uint32, send func([]byte) error, advertisements <-chan []byte, ops ipv6NetworkOperations, retryDelay time.Duration, callbacks ...func(netip.Addr) error) (cleanup func() error, resultError error) {
	iid, enabled := session.GetIPv6InterfaceID()
	if !enabled {
		return func() error { return nil }, nil
	}
	if link == nil || link.Attrs() == nil || link.Attrs().MTU < 1280 || table < firstRoutingTable || send == nil || advertisements == nil {
		return nil, fmt.Errorf("IPv6 requires a session transport, routing table, and TUN MTU of at least 1280")
	}
	binding := &ipv6Binding{link: link, table: table, vrf: ue.TunnelMode == config.TunnelVrf, session: session, ops: ops}
	for _, callback := range callbacks {
		if callback != nil {
			binding.allocationCallbacks = append(binding.allocationCallbacks, callback)
		}
	}
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
	// Keep an unavailable IPv6 router from falling through to host main routes.
	binding.claim = &netlink.Route{Family: netlink.FAMILY_V6, Dst: &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}, Table: int(table), Priority: routingTableClaimPriority, Protocol: 4, Type: unix.RTN_BLACKHOLE, Scope: netlink.SCOPE_UNIVERSE}
	if err := ops.routeAdd(binding.claim); err != nil {
		binding.claim = nil
		return cleanup, fmt.Errorf("reserve IPv6 session routing: %w", err)
	}
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
				if candidate, err := ipv6.ParseAdvertisement(packet, iid); err == nil && candidate.ValidLifetime > 0 {
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
	if _, err := binding.install(advertisement); err != nil {
		return cleanup, err
	}
	started = true
	go func() {
		defer close(done)
		defer func() {
			if len(binding.allocationCallbacks) > 0 {
				if err := binding.withdraw(); err != nil {
					log.Warn("[UE][GTP] IPv6 forwarding shutdown failed: ", err)
				}
			}
		}()
		validity := time.NewTimer(time.Duration(advertisement.ValidLifetime) * time.Second)
		defer validity.Stop()
		router := time.NewTimer(time.Duration(advertisement.RouterLifetime) * time.Second)
		defer router.Stop()
		if advertisement.RouterLifetime == 0 {
			router.Stop()
		}
		refresh := time.NewTimer(refreshDelay(advertisement))
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
				if err != nil && !errors.Is(err, ipv6.ErrNoAutonomousPrefix) {
					continue
				}
				if !candidate.Address.IsValid() {
					if allocation := session.GetIPv6(); allocation.IsValid() {
						if err := binding.updateRouter(allocation, candidate.RouterLifetime); err != nil {
							log.Warn("[UE][GTP] IPv6 router update failed: ", err)
						}
						if candidate.RouterLifetime > 0 {
							router.Reset(time.Duration(candidate.RouterLifetime) * time.Second)
						} else {
							router.Stop()
						}
					}
					continue
				}
				if candidate.ValidLifetime == 0 {
					allocation := session.GetIPv6()
					if allocation.IsValid() && candidate.Prefix.Contains(allocation) {
						if err := binding.withdraw(); err != nil {
							log.Warn("[UE][GTP] Withdrawn IPv6 prefix cleanup failed: ", err)
						}
						validity.Stop()
						router.Stop()
						refresh.Reset(retryDelay)
					}
					continue
				}
				committed, err := binding.install(candidate)
				if err != nil {
					log.Warn("[UE][GTP] IPv6 advertisement update failed: ", err)
				}
				if !committed {
					continue
				}
				// Retirement failures cannot preserve the previous allocation's timers.
				validity.Reset(time.Duration(candidate.ValidLifetime) * time.Second)
				if candidate.RouterLifetime > 0 {
					router.Reset(time.Duration(candidate.RouterLifetime) * time.Second)
				} else {
					router.Stop()
				}
				refresh.Reset(refreshDelay(candidate))
			case <-refresh.C:
				if err := send(ipv6.RouterSolicitation(iid)); err != nil {
					log.Warn("[UE][GTP] IPv6 prefix refresh solicitation failed: ", err)
				}
				refresh.Reset(retryDelay)
			case <-router.C:
				if err := binding.withdrawRouter(); err != nil {
					log.Warn("[UE][GTP] Expired IPv6 router cleanup failed: ", err)
				}
			case <-validity.C:
				if err := binding.withdraw(); err != nil {
					log.Warn("[UE][GTP] Expired IPv6 prefix cleanup failed: ", err)
				}
				router.Stop()
			}
		}
	}()
	return cleanup, nil
}
func refreshDelay(a ipv6.Advertisement) time.Duration {
	seconds := a.ValidLifetime
	if a.RouterLifetime > 0 && uint32(a.RouterLifetime) < seconds {
		seconds = uint32(a.RouterLifetime)
	}
	return time.Duration(seconds) * time.Second / 2
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
