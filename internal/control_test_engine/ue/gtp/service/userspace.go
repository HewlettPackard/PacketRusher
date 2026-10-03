// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"sync"
)

var makeUserspaceTUN = userspace.NewTUN
var userspaceRegistry = userspace.DefaultRegistry

// userspaceTunnel owns its stable UE endpoint and routing objects; the shared
// registry owns each N3 socket while any session holds it. Handover changes the
// transport binding, never the interface an application binds to.
type userspaceTunnel struct {
	port        userspace.PacketPort
	session     *userspace.Session
	link        netlink.Link
	rule        *netlink.Rule
	route       *netlink.Route
	vrf         *netlink.Vrf
	table       *routingTableReservation
	once        sync.Once
	ipv6Cleanup func() error
}

func (t *userspaceTunnel) release() {
	t.once.Do(func() {
		routingRemoved := true
		if t.ipv6Cleanup != nil {
			routingRemoved = t.ipv6Cleanup() == nil
		}
		if t.route != nil {
			routingRemoved = routingObjectRemoved(routeDel(t.route)) && routingRemoved
		}
		if t.rule != nil {
			routingRemoved = routingObjectRemoved(ruleDel(t.rule)) && routingRemoved
		}
		if t.session != nil {
			t.session.Close()
		} else if t.port != nil {
			t.port.Close()
		}
		if t.vrf != nil {
			routingRemoved = routingObjectRemoved(deleteTunnelLink(t.vrf)) && routingRemoved
		}
		if routingRemoved {
			t.table.release()
		} else {
			t.table.quarantine()
			log.Warn("[UE][GTP] Keeping routing table after incomplete userspace cleanup")
		}
	})
}
func userspaceConfig(pdu *gnbContext.GnbPDUSession, ip netip.Addr, ueIP string, ipv6IDs ...[8]byte) (userspace.Config, error) {
	remote, err := netip.ParseAddr(pdu.GetUpfIp())
	if err != nil {
		return userspace.Config{}, err
	}
	address, err := netip.ParseAddr(ueIP)
	if err != nil && (ueIP != "" || len(ipv6IDs) == 0) {
		return userspace.Config{}, fmt.Errorf("userspace UE IPv4 address: %w", err)
	}
	if pdu.GetQosId() < 0 || pdu.GetQosId() > 63 {
		return userspace.Config{}, errors.New("QFI must be between 0 and 63")
	}
	cfg := userspace.Config{Local: ip, Remote: netip.AddrPortFrom(remote, 2152), UplinkTEID: pdu.GetTeidUplink(), DownlinkTEID: pdu.GetTeidDownlink(), QFI: uint8(pdu.GetQosId()), IPv4: address}
	if len(ipv6IDs) > 0 {
		cfg.AllowIPv6 = true
		cfg.IPv6InterfaceID = ipv6IDs[0]
	}
	return cfg, nil
}
func userspaceSessionConfig(gnbPDU *gnbContext.GnbPDUSession, local netip.Addr, pdu *context.UEPDUSession) (userspace.Config, error) {
	if iid, enabled := pdu.GetIPv6InterfaceID(); enabled {
		cfg, err := userspaceConfig(gnbPDU, local, pdu.GetIp(), iid)
		if err != nil {
			return cfg, err
		}
		cfg.IPv6PrefixAllowed = func(address netip.Addr) bool {
			allocation := pdu.GetIPv6()
			return allocation.IsValid() && netip.PrefixFrom(allocation, 64).Contains(address)
		}
		return cfg, nil
	}
	return userspaceConfig(gnbPDU, local, pdu.GetIp())
}
func setUserspaceTunnelMTU(link netlink.Link, local netip.Addr, configured int, ipv6 bool) error {
	if ipv6 {
		return gtp.SetIPv6TunnelMTU(link, local, configured)
	}
	return setTunnelMTU(link, local, configured)
}
func setupUserspaceTunnel(ue *context.UEContext, pdu *context.UEPDUSession, gnbPDU *gnbContext.GnbPDUSession, local netip.Addr) error {
	cfg, err := userspaceSessionConfig(gnbPDU, local, pdu)
	if err != nil {
		return err
	}
	if pdu.GetTunInterface() != nil {
		return pdu.UpdateTunnel(gnbPDU, local)
	}
	t := &userspaceTunnel{}
	committed := false
	defer func() {
		if !committed {
			t.release()
		}
	}()
	t.port, t.link, err = makeUserspaceTUN(fmt.Sprintf("val%s", ue.GetMsin()))
	if err != nil {
		return err
	}
	if err = setUserspaceTunnelMTU(t.link, local, ue.TunnelMTU, cfg.AllowIPv6); err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: net.ParseIP(pdu.GetIp()).To4(), Mask: net.CIDRMask(32, 32)}}
	if cfg.IPv4.IsValid() {
		if err = addTunnelAddress(t.link, addr); err != nil {
			return err
		}
	}
	t.table, _, err = sessionRoutingTables.reserve(pdu)
	if err != nil {
		return err
	}
	if ue.TunnelMode == config.TunnelVrf {
		t.vrf = &netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: fmt.Sprintf("vrf%s", ue.GetMsin())}, Table: t.table.table}
		if err = addTunnelVRF(t.vrf); err != nil {
			t.vrf = nil
			return err
		}
		if err = setTunnelMaster(t.link, t.vrf); err != nil {
			return err
		}
		if err = setTunnelUp(t.vrf); err != nil {
			return err
		}
	} else if cfg.IPv4.IsValid() {
		rule := netlink.NewRule()
		rule.Priority = 100
		rule.Table = int(t.table.table)
		rule.Src = addr.IPNet
		if err = addTunnelRule(rule); err != nil {
			return err
		}
		t.rule = rule
	}
	if err = setTunnelUp(t.link); err != nil {
		return err
	}
	t.session, err = userspaceRegistry.Open(t.port, cfg)
	if err != nil {
		return err
	}
	if cfg.IPv4.IsValid() {
		route := &netlink.Route{Dst: &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, LinkIndex: t.link.Attrs().Index, Scope: netlink.SCOPE_LINK, Protocol: 4, Priority: 1, Src: addr.IP, Table: int(t.table.table)}
		if err = replaceTunnelRoute(route); err != nil {
			return err
		}
		t.route = route
	}
	if cfg.AllowIPv6 {
		t.ipv6Cleanup, err = SetupIPv6Session(ue, pdu, t.link, t.table.table, t.session.Send, t.session.Advertisements())
		if err != nil {
			return err
		}
	}
	pdu.SetTunInterface(t.link)
	pdu.SetUEInterface(t.link)
	pdu.SetTunRule(t.rule)
	pdu.SetTunRoute(t.route)
	pdu.SetVrfDevice(t.vrf)
	pdu.SetTunnelCleanup(func(bool) {
		t.release()
		pdu.SetTunInterface(nil)
		pdu.SetUEInterface(nil)
		pdu.SetTunRule(nil)
		pdu.SetTunRoute(nil)
		pdu.SetVrfDevice(nil)
	})
	pdu.SetTunnelUpdate(func(next *gnbContext.GnbPDUSession, ip netip.Addr) error {
		cfg, err := userspaceSessionConfig(next, ip, pdu)
		if err != nil {
			return err
		}
		oldMTU := t.link.Attrs().MTU
		if err = setUserspaceTunnelMTU(t.link, ip, ue.TunnelMTU, cfg.AllowIPv6); err != nil {
			return err
		}
		if err = t.session.Update(cfg); err != nil {
			if restoreErr := restoreTunnelMTU(t.link, oldMTU); restoreErr != nil {
				return fmt.Errorf("%w: %v (restore MTU: %v)", errTunnelRollback, err, restoreErr)
			}
			return err
		}
		return nil
	})
	committed = true
	log.Infof("[UE][GTP] Userspace tunnel %s configured; IPv4 %s, IPv6 %s", t.link.Attrs().Name, pdu.GetIp(), pdu.GetIPv6())
	return nil
}
