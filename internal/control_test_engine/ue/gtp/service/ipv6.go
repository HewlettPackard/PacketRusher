/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"errors"
	"net"
	"net/netip"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

var errNoIPv6 = errors.New("IPv6 user plane needs the userspace tunnel backend")

// setupIPv6 adds IPv6 to the tunnel that plumbTunnel is about to commit on link.
// The SMF only allocates an interface identifier: the UE has to ask the UPF for
// its /64 prefix (TS 23.501 §5.8.2.2.3). The answer is awaited in a goroutine, so
// that it neither blocks the UE nor costs it the IPv4 half of the tunnel; it ends
// with the tunnel or the UE at the latest. After a handover the address is known
// and moves to the new device at once.
func setupIPv6(ue *context.UEContext, pduSession *context.UEPDUSession, dedicated *dedicatedTunnel, link netlink.Link, table int) {
	address, msin := pduSession.GetIPv6(), ue.GetMsin()
	if dedicated == nil {
		log.Warn("[UE][GTP] No IPv6 for UE ", msin, ": ", errNoIPv6)
		return
	}
	plumb := func(global netip.Addr) {
		if err := dedicated.plumbIPv6(ue.TunnelMode, link, table, global); err != nil {
			log.Warn("[UE][GTP] Unable to configure IPv6 for UE ", msin, ": ", err)
			return
		}
		logTunnel(ue, link.Attrs().Name, global.String())
	}
	if address.IsGlobalUnicast() {
		plumb(address)
		return
	}
	go func() {
		global, err := dedicated.datapath.solicit(address)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) { // Otherwise released meanwhile.
				log.Warn("[UE][GTP] No IPv6 for UE ", msin, ": ", err)
			}
			return
		}
		ue.RunOnUE(func() {
			if pduSession.GetTunInterface() != link {
				return // Released or handed over meanwhile: the new tunnel solicits.
			}
			pduSession.SetIPv6(global)
			// The same tunnel, which from now on carries the packets of the prefix.
			if err := pduSession.UpdateTunnel(pduSession.GnbPduSession, pduSession.GetGnbIp()); err != nil {
				log.Warn("[UE][GTP] Unable to configure IPv6 for UE ", msin, ": ", err)
				return
			}
			plumb(global)
		})
	}()
}

// plumbIPv6 does for the UE's IPv6 address what plumbTunnel does for its IPv4 one:
// the address on the UE's endpoint, the rule sending its /64 to the UE's routing
// table unless a VRF does, and the default route through link. A handover repeats
// this for its new device: the address and the rule are then already there.
func (t *dedicatedTunnel) plumbIPv6(mode config.TunnelMode, link netlink.Link, table int, address netip.Addr) error {
	// No duplicate address detection: the prefix belongs to this UE alone.
	ip := &netlink.Addr{IPNet: &net.IPNet{IP: address.AsSlice(), Mask: net.CIDRMask(128, 128)}, Flags: syscall.IFA_F_NODAD}
	if err := addTunnelAddress(t.endpoint, ip); err != nil && !errors.Is(err, syscall.EEXIST) {
		return err
	}
	if mode != config.TunnelVrf {
		prefix := netip.PrefixFrom(address, 64).Masked().Addr()
		rule := sourceRule(&net.IPNet{IP: prefix.AsSlice(), Mask: net.CIDRMask(64, 128)}, table)
		if err := addTunnelRule(rule); err != nil && !errors.Is(err, syscall.EEXIST) {
			return err
		}
		t.rule6 = rule
	}
	return replaceTunnelRoute(defaultRoute(link, address.AsSlice(), table))
}
