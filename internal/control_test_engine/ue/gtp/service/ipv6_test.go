// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	ipv6 "my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"
	"net/netip"
	"sync"
	"testing"
	"time"
)

const raWire = "6000000000303afffe800000000000000000000000000001ff02000000000000000000000000000186009c3d400007080000000000000000030440c000000e10000007080000000020010db8123400000000000000000000"

func ipv6TestSession(t *testing.T) (*context.UEContext, *context.UEPDUSession) {
	t.Helper()
	ue := &context.UEContext{PDUSessionType: config.PDUSessionType(2), TunnelMode: config.TunnelTun}
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	require.NoError(t, session.SetPDUAddress(2, &ie.PDUAddr{IPv6IfId: []byte{0, 0, 0, 0, 0, 0, 0, 7}}))
	return ue, session
}

func TestIPv6PrefixRenewalRenumberingExpiryAndRediscovery(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 8)
	var mu sync.Mutex
	addresses := make(map[string]bool)
	var renewals int
	ops := ipv6NetworkOperations{
		addressAdd: func(_ netlink.Link, address *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			addresses[address.IP.String()] = true
			return nil
		},
		addressReplace: func(_ netlink.Link, address *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			renewals++
			addresses[address.IP.String()] = true
			return nil
		},
		addressDel: func(_ netlink.Link, address *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			delete(addresses, address.IP.String())
			return nil
		},
		ruleAdd: func(*netlink.Rule) error { return nil }, ruleDel: func(*netlink.Rule) error { return nil },
		routeAdd: func(*netlink.Route) error { return nil }, routeDel: func(*netlink.Route) error { return nil },
	}
	first, _ := hex.DecodeString(raWire)
	advertisements <- first
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, ops, 10*time.Millisecond)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cleanup()) })
	advertisements <- first
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return renewals == 1 }, time.Second, time.Millisecond)
	second := append([]byte(nil), first...)
	second[76], second[77] = 0x56, 0x78
	binary.BigEndian.PutUint32(second[60:64], 1)
	binary.BigEndian.PutUint32(second[64:68], 1)
	second[42], second[43] = 0, 0
	checksum := udp6Checksum(second)
	binary.BigEndian.PutUint16(second[42:44], checksum)
	advertisements <- second
	require.Eventually(t, func() bool { return session.GetIPv6() == netip.MustParseAddr("2001:db8:5678::7") }, time.Second, time.Millisecond)
	mu.Lock()
	require.False(t, addresses["2001:db8:1234::7"], "old prefix must be retired after the new route commits")
	mu.Unlock()
	require.Eventually(t, func() bool { return !session.GetIPv6().IsValid() }, 1500*time.Millisecond, time.Millisecond)
	mu.Lock()
	require.True(t, addresses["fe80::7"])
	require.False(t, addresses["2001:db8:5678::7"])
	mu.Unlock()
	advertisements <- first
	require.Eventually(t, func() bool { return session.GetIPv6() == netip.MustParseAddr("2001:db8:1234::7") }, time.Second, time.Millisecond)
	withdrawal := append([]byte(nil), first...)
	binary.BigEndian.PutUint32(withdrawal[60:64], 0)
	binary.BigEndian.PutUint32(withdrawal[64:68], 0)
	withdrawal[42], withdrawal[43] = 0, 0
	binary.BigEndian.PutUint16(withdrawal[42:44], udp6Checksum(withdrawal))
	advertisements <- withdrawal
	require.Eventually(t, func() bool { return !session.GetIPv6().IsValid() }, time.Second, time.Millisecond)
	require.NoError(t, cleanup())
	mu.Lock()
	require.Empty(t, addresses)
	mu.Unlock()
}

func TestIPv6SessionDiscoverySourcePolicyAndCleanup(t *testing.T) {
	for _, mode := range []config.TunnelMode{config.TunnelTun, config.TunnelVrf} {
		ue, session := ipv6TestSession(t)
		ue.TunnelMode = mode
		link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: "val7", Index: 77, MTU: 1456}}
		advertisements := make(chan []byte, 2)
		var events []string
		ops := ipv6NetworkOperations{
			addressAdd: func(_ netlink.Link, a *netlink.Addr) error {
				events = append(events, "address+"+a.IP.String())
				return nil
			},
			addressDel: func(_ netlink.Link, a *netlink.Addr) error {
				events = append(events, "address-"+a.IP.String())
				return nil
			},
			ruleAdd: func(r *netlink.Rule) error {
				require.Equal(t, netlink.FAMILY_V6, r.Family)
				require.Equal(t, 1000, r.Table)
				require.Equal(t, "2001:db8:1234::7/128", r.Src.String())
				events = append(events, "rule+")
				return nil
			},
			ruleDel: func(*netlink.Rule) error { events = append(events, "rule-"); return nil },
			routeAdd: func(r *netlink.Route) error {
				if r.Type == unix.RTN_BLACKHOLE {
					events = append(events, "claim+")
					return nil
				}
				require.Equal(t, "::/0", r.Dst.String())
				require.Equal(t, 77, r.LinkIndex)
				require.Equal(t, 1000, r.Table)
				require.Equal(t, "2001:db8:1234::7", r.Src.String())
				events = append(events, "route+")
				return nil
			},
			routeDel: func(r *netlink.Route) error {
				if r.Type == unix.RTN_BLACKHOLE {
					events = append(events, "claim-")
				} else {
					events = append(events, "route-")
				}
				return nil
			},
		}
		send := func(packet []byte) error {
			require.Equal(t, byte(133), packet[40])
			invalid, _ := hex.DecodeString(raWire)
			invalid[7] = 64
			advertisements <- invalid
			valid, _ := hex.DecodeString(raWire)
			advertisements <- valid
			return nil
		}
		cleanup, err := setupIPv6Session(ue, session, link, 1000, send, advertisements, ops, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, netip.MustParseAddr("2001:db8:1234::7"), session.GetIPv6())
		require.NoError(t, cleanup())
		require.NoError(t, cleanup())
		if mode == config.TunnelVrf {
			require.Equal(t, []string{"address+fe80::7", "claim+", "address+2001:db8:1234::7", "route+", "route-", "address-2001:db8:1234::7", "address-fe80::7", "claim-"}, events)
		} else {
			require.Equal(t, []string{"address+fe80::7", "claim+", "rule+", "address+2001:db8:1234::7", "route+", "route-", "address-2001:db8:1234::7", "rule-", "address-fe80::7", "claim-"}, events)
		}
	}
}

func TestIPv6SessionDiscoveryTimeoutLeavesNoGlobalRoute(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	var sent, removed int
	ops := ipv6NetworkOperations{addressAdd: func(netlink.Link, *netlink.Addr) error { return nil }, addressDel: func(netlink.Link, *netlink.Addr) error { removed++; return nil }, routeAdd: func(*netlink.Route) error { return nil }, routeDel: func(*netlink.Route) error { return nil }}
	_, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { sent++; return nil }, make(chan []byte), ops, time.Millisecond)
	require.ErrorContains(t, err, "three solicitations")
	require.Equal(t, 3, sent)
	require.Equal(t, 1, removed)
	require.False(t, session.GetIPv6().IsValid())
	link.Attrs().MTU = 1279
	_, err = setupIPv6Session(ue, session, link, 1000, func([]byte) error { return errors.New("unexpected") }, make(chan []byte), ops, time.Millisecond)
	require.ErrorContains(t, err, "1280")
	require.Equal(t, 3, sent)
}

func TestIPv6RouterExpiryPreservesAddressAndBlocksHostFallback(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 8)
	var mu sync.Mutex
	addressPresent, policyPresent, routerPresent, claimPresent := false, false, false, false
	ops := ipv6NetworkOperations{
		addressAdd: func(_ netlink.Link, a *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			if !a.IP.IsLinkLocalUnicast() {
				addressPresent = true
			}
			return nil
		},
		addressReplace: func(netlink.Link, *netlink.Addr) error { return nil },
		addressDel: func(_ netlink.Link, a *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			if !a.IP.IsLinkLocalUnicast() {
				addressPresent = false
			}
			return nil
		},
		ruleAdd: func(*netlink.Rule) error { mu.Lock(); defer mu.Unlock(); policyPresent = true; return nil },
		ruleDel: func(*netlink.Rule) error { mu.Lock(); defer mu.Unlock(); policyPresent = false; return nil },
		routeAdd: func(r *netlink.Route) error {
			mu.Lock()
			defer mu.Unlock()
			if r.Type == unix.RTN_BLACKHOLE {
				claimPresent = true
			} else {
				routerPresent = true
			}
			return nil
		},
		routeDel: func(r *netlink.Route) error {
			mu.Lock()
			defer mu.Unlock()
			if r.Type == unix.RTN_BLACKHOLE {
				claimPresent = false
			} else {
				routerPresent = false
			}
			return nil
		},
	}
	first, _ := hex.DecodeString(raWire)
	binary.BigEndian.PutUint16(first[46:48], 1)
	first[42], first[43] = 0, 0
	binary.BigEndian.PutUint16(first[42:44], udp6Checksum(first))
	advertisements <- first
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, ops, 10*time.Millisecond)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cleanup()) })
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return !routerPresent }, 1500*time.Millisecond, time.Millisecond)
	require.True(t, session.GetIPv6().IsValid(), "router expiry must preserve the still-valid autonomous address")
	mu.Lock()
	require.True(t, addressPresent)
	require.True(t, policyPresent)
	require.True(t, claimPresent, "missing router must terminate table lookup before host main routes")
	mu.Unlock()
	zero := append([]byte(nil), first...)
	zero[46], zero[47] = 0, 0
	zero[42], zero[43] = 0, 0
	binary.BigEndian.PutUint16(zero[42:44], udp6Checksum(zero))
	advertisements <- zero
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return !routerPresent && addressPresent && claimPresent }, time.Second, time.Millisecond)
	renewed, _ := hex.DecodeString(raWire)
	advertisements <- renewed
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return routerPresent }, time.Second, time.Millisecond)
	require.True(t, session.GetIPv6().IsValid())
	// A default-router withdrawal need not contain a Prefix Information option.
	noPrefix := append([]byte(nil), zero[:56]...)
	binary.BigEndian.PutUint16(noPrefix[4:6], 16)
	noPrefix[42], noPrefix[43] = 0, 0
	binary.BigEndian.PutUint16(noPrefix[42:44], udp6Checksum(noPrefix))
	advertisements <- noPrefix
	require.Eventually(t, func() bool { mu.Lock(); defer mu.Unlock(); return !routerPresent }, time.Second, time.Millisecond)
	require.True(t, session.GetIPv6().IsValid(), "RA without PIO must not reset or withdraw the address lease")
}

func TestIPv6CommittedRenumberRetirementFailureUsesNewValidity(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 8)
	var mu sync.Mutex
	retirementFails := true
	policies := make(map[string]bool)
	ops := ipv6NetworkOperations{
		addressAdd: func(netlink.Link, *netlink.Addr) error { return nil }, addressReplace: func(netlink.Link, *netlink.Addr) error { return nil },
		addressDel: func(_ netlink.Link, a *netlink.Addr) error {
			mu.Lock()
			defer mu.Unlock()
			if a.IP.String() == "2001:db8:1234::7" && retirementFails {
				return errors.New("old address busy")
			}
			return nil
		},
		ruleAdd: func(rule *netlink.Rule) error {
			mu.Lock()
			defer mu.Unlock()
			policies[rule.Src.String()] = true
			return nil
		},
		ruleDel: func(rule *netlink.Rule) error {
			mu.Lock()
			defer mu.Unlock()
			delete(policies, rule.Src.String())
			return nil
		},
		routeAdd: func(*netlink.Route) error { return nil }, routeDel: func(*netlink.Route) error { return nil },
	}
	first, _ := hex.DecodeString(raWire)
	binary.BigEndian.PutUint32(first[60:64], 1)
	binary.BigEndian.PutUint32(first[64:68], 1)
	first[42], first[43] = 0, 0
	binary.BigEndian.PutUint16(first[42:44], udp6Checksum(first))
	advertisements <- first
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, ops, 10*time.Millisecond)
	require.NoError(t, err)
	t.Cleanup(func() { mu.Lock(); retirementFails = false; mu.Unlock(); require.NoError(t, cleanup()) })
	second, _ := hex.DecodeString(raWire)
	second[76], second[77] = 0x56, 0x78
	binary.BigEndian.PutUint32(second[60:64], 3)
	binary.BigEndian.PutUint32(second[64:68], 2)
	second[42], second[43] = 0, 0
	binary.BigEndian.PutUint16(second[42:44], udp6Checksum(second))
	advertisements <- second
	allocation := netip.MustParseAddr("2001:db8:5678::7")
	require.Eventually(t, func() bool { return session.GetIPv6() == allocation }, time.Second, time.Millisecond)
	// The previous allocation expires after one second; the committed new lease
	// remains valid for three even when retirement needs a later cleanup retry.
	time.Sleep(1200 * time.Millisecond)
	require.Equal(t, allocation, session.GetIPv6())
	mu.Lock()
	require.True(t, policies["2001:db8:1234::7/128"], "orphaned address must retain its source policy until removal succeeds")
	retirementFails = false
	mu.Unlock()
	require.NoError(t, cleanup())
	mu.Lock()
	require.Empty(t, policies)
	mu.Unlock()
}

func TestIPv6InitialZeroRouterKeepsAddressAndBlockingPolicy(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	packet, _ := hex.DecodeString(raWire)
	packet[46], packet[47], packet[42], packet[43] = 0, 0, 0, 0
	binary.BigEndian.PutUint16(packet[42:44], udp6Checksum(packet))
	advertisements <- packet
	var routes []*netlink.Route
	var policy bool
	ops := ipv6NetworkOperations{
		addressAdd: func(netlink.Link, *netlink.Addr) error { return nil }, addressDel: func(netlink.Link, *netlink.Addr) error { return nil },
		ruleAdd: func(*netlink.Rule) error { policy = true; return nil }, ruleDel: func(*netlink.Rule) error { policy = false; return nil },
		routeAdd: func(route *netlink.Route) error { routes = append(routes, route); return nil }, routeDel: func(*netlink.Route) error { return nil },
	}
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, ops, time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, "2001:db8:1234::7", session.GetIPv6().String())
	require.True(t, policy)
	require.Len(t, routes, 1)
	require.Equal(t, unix.RTN_BLACKHOLE, routes[0].Type)
	require.NoError(t, cleanup())
	require.False(t, policy)
}

func TestIPv6WithdrawalFailureRetainsSourcePolicyAndClaim(t *testing.T) {
	_, session := ipv6TestSession(t)
	var addressBusy = true
	var policy, claim bool
	ops := ipv6NetworkOperations{
		addressAdd: func(netlink.Link, *netlink.Addr) error { return nil },
		addressDel: func(netlink.Link, *netlink.Addr) error {
			if addressBusy {
				return errors.New("address busy")
			}
			return nil
		},
		ruleAdd: func(*netlink.Rule) error { policy = true; return nil }, ruleDel: func(*netlink.Rule) error { policy = false; return nil },
		routeAdd: func(*netlink.Route) error { return nil }, routeDel: func(route *netlink.Route) error {
			if route.Type == unix.RTN_BLACKHOLE {
				claim = false
			}
			return nil
		},
	}
	binding := &ipv6Binding{link: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77}}, table: 1000, session: session, ops: ops, claim: &netlink.Route{Type: unix.RTN_BLACKHOLE}}
	claim = true
	packet, _ := hex.DecodeString(raWire)
	advertisement, err := ipv6.ParseAdvertisement(packet, [8]byte{0, 0, 0, 0, 0, 0, 0, 7})
	require.NoError(t, err)
	committed, err := binding.install(advertisement)
	require.True(t, committed)
	require.NoError(t, err)
	require.Error(t, binding.withdraw())
	require.False(t, session.GetIPv6().IsValid())
	require.True(t, policy)
	require.Error(t, binding.cleanup())
	require.True(t, policy, "a remaining address must stay in its session table")
	require.True(t, claim, "quarantined routing must keep a blocking default")
	addressBusy = false
	require.NoError(t, binding.cleanup())
	require.False(t, policy)
	require.False(t, claim)
}

func TestIPv6StagingPolicyFailureCannotAssignUnisolatedAddress(t *testing.T) {
	_, session := ipv6TestSession(t)
	var addressAdds, policyDeletes int
	policyAddFails, policyDeleteFails := true, true
	ops := ipv6NetworkOperations{
		addressAdd: func(netlink.Link, *netlink.Addr) error { addressAdds++; return errors.New("address rejected") },
		ruleAdd: func(*netlink.Rule) error {
			if policyAddFails {
				return errors.New("policy rejected")
			}
			return nil
		},
		ruleDel: func(*netlink.Rule) error {
			policyDeletes++
			if policyDeleteFails {
				return errors.New("policy busy")
			}
			return nil
		},
	}
	binding := &ipv6Binding{link: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77}}, table: 1000, session: session, ops: ops}
	packet, _ := hex.DecodeString(raWire)
	advertisement, err := ipv6.ParseAdvertisement(packet, [8]byte{0, 0, 0, 0, 0, 0, 0, 7})
	require.NoError(t, err)
	committed, err := binding.install(advertisement)
	require.False(t, committed)
	require.ErrorContains(t, err, "policy rejected")
	require.Zero(t, addressAdds, "address must not become usable before its source policy exists")
	require.Empty(t, binding.addresses)
	require.Empty(t, binding.rules)
	policyAddFails = false
	committed, err = binding.install(advertisement)
	require.False(t, committed)
	require.ErrorContains(t, err, "address rejected")
	require.ErrorContains(t, err, "policy busy")
	require.Equal(t, 1, addressAdds)
	require.Equal(t, 1, policyDeletes)
	require.Empty(t, binding.addresses)
	require.Len(t, binding.rules, 1, "failed policy rollback must remain tracked")
	policyDeleteFails = false
	require.NoError(t, binding.cleanup())
	require.Empty(t, binding.rules)
}
