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
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/context"
)

var errNoIPv6 = errors.New("IPv6 user plane needs the userspace tunnel backend")

// solicitIPv6 adds IPv6 to the tunnel. The SMF only allocates an interface
// identifier: the UE has to ask the UPF for its /64 prefix (TS 23.501 §5.8.2.2.3).
// The answer is awaited in a goroutine, so that it neither blocks the UE nor costs
// it the IPv4 half of the tunnel; it ends with the datapath at the latest.
func (t *tunnel) solicitIPv6(ue *context.UEContext, pduSession *context.UEPDUSession) {
	d, linkLocal, msin := t.datapath, pduSession.GetIPv6(), ue.GetMsin()
	go func() {
		global, err := d.solicit(linkLocal)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) { // Otherwise released or handed over meanwhile.
				log.Warn("[UE][GTP] No IPv6 for UE ", msin, ": ", err)
			}
			return
		}
		ue.RunOnUE(func() {
			if pduSession.Tunnel != t || t.datapath != d {
				return // Released or handed over meanwhile: the new datapath solicits.
			}
			pduSession.SetIPv6(global)
			// The same datapath, which from now on carries the packets of the prefix.
			err := d.refresh(pduSession.GnbPduSession, pduSession.GetIp(), t.gnbIP)
			if err == nil {
				err = t.plumbIPv6(global)
			}
			if err != nil {
				log.Warn("[UE][GTP] Unable to configure IPv6 for UE ", msin, ": ", err)
				return
			}
			logTunnel(ue, t.link.Attrs().Name, global.String())
		})
	}()
}

// plumbIPv6 does for the UE's IPv6 address what create does for its IPv4 one: the
// address on val<MSIN>, the rule sending its /64 to the UE's routing table, and
// the default route, which then follows the UE as the IPv4 one does.
func (t *tunnel) plumbIPv6(address netip.Addr) error {
	// No duplicate address detection: the prefix belongs to this UE alone.
	ip := &netlink.Addr{IPNet: &net.IPNet{IP: address.AsSlice(), Mask: net.CIDRMask(128, 128)}, Flags: syscall.IFA_F_NODAD}
	if err := netlink.AddrAdd(t.endpoint, ip); err != nil {
		return err
	}
	// The kernel only delivers to the address a moment later, once a work queue has
	// routed it, and drops what the UE receives until then.
	for range 1000 {
		routes, _ := netlink.RouteGetWithOptions(ip.IP, &netlink.RouteGetOptions{Iif: t.link.Attrs().Name})
		if len(routes) > 0 && routes[0].Type == syscall.RTN_LOCAL {
			break
		}
		time.Sleep(time.Millisecond)
	}
	prefix := netip.PrefixFrom(address, 64).Masked().Addr()
	if err := t.addRule(&net.IPNet{IP: prefix.AsSlice(), Mask: net.CIDRMask(64, 128)}); err != nil {
		return err
	}
	route := defaultRoute(address.AsSlice(), t.table)
	route.LinkIndex = t.link.Attrs().Index
	if err := netlink.RouteReplace(route); err != nil {
		return err
	}
	t.routes = append(t.routes, route)
	return nil
}
