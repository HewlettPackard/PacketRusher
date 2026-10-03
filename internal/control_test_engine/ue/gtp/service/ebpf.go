// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

var ebpfRegistry = ebpfgtp.DefaultRegistry
var makeEBPFTUN = userspace.NewTUN

type ebpfSession interface {
	Close() error
	Update(ebpfgtp.Config) error
	ProgramFD() int
}

// Quarantine retains the file descriptor as well as the routing reservation.
// A finalizer must not remove an endpoint still named by an active BPF entry.
var quarantinedEBPFEndpoints sync.Map

func (t *ebpfTunnel) quarantine(err error) {
	quarantinedEBPFEndpoints.Store(t, struct{}{})
	if t.table != nil {
		t.table.quarantine()
	}
	log.Error("[UE][eBPF] Quarantined endpoint after incomplete cleanup: ", err)
}

type ebpfTunnel struct {
	port    userspace.PacketPort
	session ebpfSession
	link    netlink.Link
	rule    *netlink.Rule
	route   *netlink.Route
	table   *routingTableReservation
	once    sync.Once
}

func (t *ebpfTunnel) release() {
	t.once.Do(func() {
		// Retain the endpoint/port reservation if deactivation fails, preventing a
		// stale map entry from redirecting into an interface index later reused.
		if t.session != nil {
			if err := t.session.Close(); err != nil {
				t.quarantine(err)
				return
			}
		}
		// Remove the assigned address with its TUN before removing source policy.
		if t.port != nil {
			if err := t.port.Close(); err != nil {
				t.quarantine(err)
				return
			}
		}
		removed := true
		if t.route != nil {
			removed = routingObjectRemoved(routeDel(t.route)) && removed
		}
		if t.rule != nil {
			removed = routingObjectRemoved(ruleDel(t.rule)) && removed
		}
		if t.table != nil {
			if removed {
				t.table.release()
			} else {
				t.table.quarantine()
			}
		}
	})
}
func ebpfSessionConfig(pdu *context.UEPDUSession, gnbPDU *gnbContext.GnbPDUSession, local netip.Addr, endpoint, mtu int) (ebpfgtp.Config, error) {
	if _, ipv6 := pdu.GetIPv6InterfaceID(); ipv6 {
		return ebpfgtp.Config{}, errors.New("eBPF supports IPv4 PDU sessions; IPv6/IPv4v6 requires the userspace backend")
	}
	remote, err := netip.ParseAddr(gnbPDU.GetUpfIp())
	if err != nil {
		return ebpfgtp.Config{}, err
	}
	ip, err := netip.ParseAddr(pdu.GetIp())
	if err != nil {
		return ebpfgtp.Config{}, err
	}
	if gnbPDU.GetQosId() < 0 || gnbPDU.GetQosId() > 63 {
		return ebpfgtp.Config{}, errors.New("eBPF QFI must be between 0 and 63")
	}
	cfg := ebpfgtp.Config{Local: local, Remote: remote, IPv4: ip, UplinkTEID: gnbPDU.GetTeidUplink(), DownlinkTEID: gnbPDU.GetTeidDownlink(), QFI: uint8(gnbPDU.GetQosId()), EndpointIfIndex: endpoint, MTU: mtu}
	return cfg, cfg.Validate()
}
func setupEBPFTunnel(ue *context.UEContext, pdu *context.UEPDUSession, gnbPDU *gnbContext.GnbPDUSession, local netip.Addr) error {
	if ue.TunnelMode == config.TunnelVrf {
		return errors.New("eBPF currently requires policy routing; --tunnel vrf is unsupported")
	}
	if _, ipv6 := pdu.GetIPv6InterfaceID(); ipv6 {
		return errors.New("eBPF supports IPv4 only; select userspace for IPv6/IPv4v6")
	}
	if pdu.GetTunInterface() != nil {
		return pdu.UpdateTunnel(gnbPDU, local)
	}
	t := &ebpfTunnel{}
	committed := false
	defer func() {
		if !committed {
			t.release()
		}
	}()
	var err error
	t.port, t.link, err = makeEBPFTUN(fmt.Sprintf("val%s", ue.GetMsin()))
	if err != nil {
		return err
	}
	if err = setTunnelMTU(t.link, local, ue.TunnelMTU); err != nil {
		return err
	}
	if err = ebpfgtp.ConfigureEndpoint(t.link); err != nil {
		return err
	}
	cfg, err := ebpfSessionConfig(pdu, gnbPDU, local, t.link.Attrs().Index, t.link.Attrs().MTU)
	if err != nil {
		return err
	}
	t.table, _, err = sessionRoutingTables.reserve(pdu)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: net.IP(cfg.IPv4.AsSlice()), Mask: net.CIDRMask(32, 32)}}
	rule := netlink.NewRule()
	rule.Priority = 100
	rule.Table = int(t.table.table)
	rule.Src = addr.IPNet
	if err = addTunnelRule(rule); err != nil {
		return err
	}
	t.rule = rule
	if err = addTunnelAddress(t.link, addr); err != nil {
		return err
	}
	if err = setTunnelUp(t.link); err != nil {
		return err
	}
	ebpfRegistry.SetWarningHandler(func(err error) { log.Warn("[UE][eBPF] ", err) })
	t.session, err = ebpfRegistry.Open(cfg)
	if err != nil {
		return err
	}
	encap := &netlink.BpfEncap{}
	if err = encap.SetProg(nl.LWT_BPF_XMIT, t.session.ProgramFD(), "packetrusher_gtpu"); err != nil {
		return err
	}
	if err = encap.SetXmitHeadroom(44); err != nil {
		return err
	}
	route := &netlink.Route{Dst: &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, LinkIndex: t.link.Attrs().Index, Scope: netlink.SCOPE_LINK, Protocol: 4, Priority: 1, Src: addr.IP, Table: int(t.table.table), Encap: encap}
	if err = replaceTunnelRoute(route); err != nil {
		return fmt.Errorf("install eBPF LWT route: %w", err)
	}
	t.route = route
	pdu.SetTunInterface(t.link)
	pdu.SetUEInterface(t.link)
	pdu.SetTunRule(t.rule)
	pdu.SetTunRoute(t.route)
	pdu.SetTunnelCleanup(func(bool) { t.release() })
	pdu.SetTunnelUpdate(func(next *gnbContext.GnbPDUSession, ip netip.Addr) error {
		oldMTU := t.link.Attrs().MTU
		if err := setTunnelMTU(t.link, ip, ue.TunnelMTU); err != nil {
			return err
		}
		cfg, err := ebpfSessionConfig(pdu, next, ip, t.link.Attrs().Index, t.link.Attrs().MTU)
		if err == nil {
			err = t.session.Update(cfg)
		}
		if err != nil {
			if restoreErr := restoreTunnelMTU(t.link, oldMTU); restoreErr != nil {
				return fmt.Errorf("%w: %v (restore MTU: %v)", errTunnelRollback, err, restoreErr)
			}
			return err
		}
		return nil
	})
	committed = true
	log.Infof("[UE][eBPF] IPv4 GTP-U fastpath configured on %s for %s", t.link.Attrs().Name, pdu.GetIp())
	return nil
}
