// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/hex"
	"errors"
	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"net/netip"
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
