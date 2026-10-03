// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
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
	withdrawal[46], withdrawal[47] = 0, 0
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
				require.Equal(t, "::/0", r.Dst.String())
				require.Equal(t, 77, r.LinkIndex)
				require.Equal(t, 1000, r.Table)
				require.Equal(t, "2001:db8:1234::7", r.Src.String())
				events = append(events, "route+")
				return nil
			},
			routeDel: func(*netlink.Route) error { events = append(events, "route-"); return nil },
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
			require.Equal(t, []string{"address+fe80::7", "address+2001:db8:1234::7", "route+", "route-", "address-2001:db8:1234::7", "address-fe80::7"}, events)
		} else {
			require.Equal(t, []string{"address+fe80::7", "address+2001:db8:1234::7", "rule+", "route+", "route-", "rule-", "address-2001:db8:1234::7", "address-fe80::7"}, events)
		}
	}
}

func TestIPv6SessionDiscoveryTimeoutLeavesNoGlobalRoute(t *testing.T) {
	ue, session := ipv6TestSession(t)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	var sent, removed int
	ops := ipv6NetworkOperations{addressAdd: func(netlink.Link, *netlink.Addr) error { return nil }, addressDel: func(netlink.Link, *netlink.Addr) error { removed++; return nil }}
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
