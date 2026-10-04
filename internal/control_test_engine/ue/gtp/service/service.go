/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"net/netip"
	"os"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
	log "my5G-RANTester/internal/log"
)

// pathSwitchDelay is how long the core may take to switch the downlink path to the new
// gNB once the UE has moved there.
const pathSwitchDelay = time.Second

// datapath is the backend-specific part of a tunnel: what carries the UE's packets
// between a device of this host and the UPF. Everything else is common, see tunnel.
type datapath interface {
	// refresh applies the PDU session's current TEIDs and UPF, on the same N3 address.
	refresh(pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) error
	// close removes it and returns once its N3 address can be bound again.
	close()
	// solicit asks the UPF for the UE's IPv6 prefix, from its link-local address,
	// and returns its global address.
	solicit(linkLocal netip.Addr) (netip.Addr, error)
}

// tunnel is the user plane of a PDU session on this host. The first setup of the
// session creates it and it lasts as long as the session: a handover only gives it
// another datapath, so the UE keeps its addresses, its routing and its connections.
type tunnel struct {
	gnbIP    netip.Addr       // N3 address of the UE's gNB
	datapath datapath         // to the UPF, from that address
	left     *datapath        // of the gNB the UE just left, still receiving
	link     netlink.Link     // its device: gtp0<MSIN> or gtp1<MSIN>, or the one the gNB shares
	endpoint netlink.Link     // val<MSIN>, holding the UE's addresses; the shared device does itself
	vrf      *netlink.Vrf     // vrf<MSIN>, in VRF mode
	table    int              // the UE's routing table
	rules    []*netlink.Rule  // to it, from the UE's addresses, unless the VRF leads there
	routes   []*netlink.Route // its default routes, through link
}

// SetupGtpInterface sets up the tunnel of the UE's PDU session, or updates it after
// a handover or a change of its TEIDs. A UE that gets no tunnel keeps running.
func SetupGtpInterface(ue *context.UEContext, msg gnbContext.UEMessage) {
	gnbPduSession := msg.GNBPduSessions[0]
	pduSession, err := ue.GetPduSession(uint8(gnbPduSession.GetPduSessionId()))
	if pduSession == nil || err != nil {
		log.Error("[GNB][GTP] Aborting the setup of PDU Session ", gnbPduSession.GetPduSessionId(), ", this PDU session was not succesfully configured on the UE's side.")
		return
	}
	pduSession.GnbPduSession = gnbPduSession

	if ue.TunnelMode == config.TunnelDisabled {
		log.Info(fmt.Sprintf("[UE][GTP] Interface for UE %s has not been created. Tunnel has been disabled.", ue.GetMsin()))
		return
	}

	// Bounded concurrency through the plumbing below; see setupSlots. A dedicated
	// device per UE keeps its own pacing, the 500 ms registration floor.
	if ue.TunnelMode == config.TunnelShared {
		acquireSetupSlot()
		defer releaseSetupSlot()
	}

	if pduSession.Id != 1 {
		log.Warn("[GNB][GTP] Only one tunnel per UE is supported for now, no tunnel will be created for second PDU Session of given UE")
		return
	}

	if err := setupTunnel(ue, pduSession, msg); err != nil {
		log.Error("[UE][GTP] Unable to set up the tunnel of UE ", ue.GetMsin(), ": ", err)
	}
}

// setupTunnel starts a datapath on the gNB's N3 address and moves the UE's routes to
// it: the first setup and a handover are the same, but for what the first creates for
// the whole session. A handover that fails leaves the UE its previous datapath.
func setupTunnel(ue *context.UEContext, pduSession *context.UEPDUSession, msg gnbContext.UEMessage) (err error) {
	pdu, ueIP, msin := pduSession.GnbPduSession, pduSession.GetIp(), ue.GetMsin()
	t, _ := pduSession.Tunnel.(*tunnel)

	// The datapath is bound to this N3 address already: only its TEIDs or UPF change.
	if t != nil && t.gnbIP == msg.GnbIp {
		return t.datapath.refresh(pdu, ueIP, msg.GnbIp)
	}

	if t == nil {
		t = &tunnel{}
		defer func() {
			if err != nil {
				t.Release()
			}
		}()
		if err := t.create(ue, pduSession); err != nil {
			return err
		}
	}

	t.leave(t.left)
	// A handover alternates between two names: both devices exist until the routes moved.
	name := "gtp0" + msin
	if t.link != nil && t.link.Attrs().Name == name {
		name = "gtp1" + msin
	}
	var d datapath
	switch { // The only place that depends on the tunnel backend.
	case ue.TunnelMode == config.TunnelShared:
		if d, err = startShared(msg.GtpDevice, pdu, ueIP, msg.GnbIp); err == nil {
			name = msg.GtpDevice.Name()
		}
	case ue.TunnelBackend == config.TunnelBackendEBPF:
		d, err = startEBPF(name, pdu, pduSession, msg.GnbIp)
	case ue.TunnelBackend == config.TunnelBackendUserspace:
		d, err = startUserspace(name, pdu, pduSession, msg.GnbIp)
	default:
		d, err = startGtp5g(name, pdu, ueIP, msg.GnbIp)
	}
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			d.close()
		}
	}()

	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	if t.endpoint != nil { // The gNB sets the MTU of the device its UEs share.
		// On val<MSIN> for "ip link" only: packets take the MTU of the route's device.
		if err := gtp.SetTunnelMTU(msg.GnbIp, ue.TunnelMTU, pduSession.GetIPv6().IsValid(), link, t.endpoint); err != nil {
			return fmt.Errorf("tunnel MTU: %w", err)
		}
		// With the UE's address and a loose reverse-path filter, the device still
		// receives once the routes have moved to the next gNB's.
		_ = os.WriteFile("/proc/sys/net/ipv4/conf/"+name+"/rp_filter", []byte("2"), 0)
		if v4 := net.ParseIP(ueIP).To4(); v4 != nil {
			if err := netlink.AddrReplace(link, &netlink.Addr{IPNet: &net.IPNet{IP: v4, Mask: net.CIDRMask(32, 32)}}); err != nil {
				return fmt.Errorf("UE address on %s: %w", name, err)
			}
		}
	}
	if t.vrf != nil {
		if err := netlink.LinkSetMaster(link, t.vrf); err != nil {
			return fmt.Errorf("attaching %s to its VRF: %w", name, err)
		}
	}
	for _, route := range t.routes {
		route.LinkIndex = link.Attrs().Index
		if err := netlink.RouteReplace(route); err != nil {
			return fmt.Errorf("route %s: %w", route, err)
		}
	}

	// After an N2 handover the UPF sends to the previous gNB until the core has switched
	// the path: its datapath keeps receiving for that long.
	if previous := t.datapath; previous != nil {
		left := &previous
		t.left = left
		ue.RunOnUEAfter(pathSwitchDelay, func() { t.leave(left) })
	}
	t.gnbIP, t.datapath, t.link = msg.GnbIp, d, link
	pduSession.Tunnel = t

	if ueIP != "" {
		logTunnel(ue, name, ueIP)
	}
	// The address of the prefix the UPF advertised follows the UE, as its IPv4 one.
	if ipv6 := pduSession.GetIPv6(); ipv6.IsValid() && !ipv6.IsGlobalUnicast() {
		t.solicitIPv6(ue, pduSession)
	}
	return nil
}

// create gives the UE what it keeps for the whole PDU session: its routing table and,
// leading there, either its VRF or the rule for its address, which val<MSIN> holds.
func (t *tunnel) create(ue *context.UEContext, pduSession *context.UEPDUSession) error {
	msin := ue.GetMsin()
	ueIP := net.ParseIP(pduSession.GetIp()).To4() // nil for an IPv6 PDU session
	if ue.TunnelMode == config.TunnelShared {
		if ueIP == nil {
			return errNoIPv6
		}
	} else {
		// A run that was killed left its devices behind: they would refuse the new ones.
		for _, device := range []string{"gtp0", "gtp1", "val", "vrf"} {
			removeLink(device + msin)
		}
		endpoint := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "val" + msin}}
		if err := netlink.LinkAdd(endpoint); err != nil {
			return fmt.Errorf("creating %s: %w", endpoint.Name, err)
		}
		t.endpoint = endpoint
		if err := netlink.LinkSetUp(endpoint); err != nil {
			return err
		}
		if ueIP != nil {
			if err := netlink.AddrAdd(endpoint, &netlink.Addr{IPNet: &net.IPNet{IP: ueIP, Mask: net.CIDRMask(32, 32)}}); err != nil {
				return fmt.Errorf("UE address: %w", err)
			}
		}
	}

	t.table = routingTable(ueIP, t.endpoint)
	if ue.TunnelMode == config.TunnelVrf {
		vrf := &netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: "vrf" + msin}, Table: uint32(t.table)}
		if err := netlink.LinkAdd(vrf); err != nil {
			return fmt.Errorf("creating %s: %w", vrf.Name, err)
		}
		t.vrf = vrf
		if err := netlink.LinkSetMaster(t.endpoint, vrf); err != nil {
			return fmt.Errorf("attaching %s to its VRF: %w", t.endpoint.Attrs().Name, err)
		}
		if err := netlink.LinkSetUp(vrf); err != nil {
			return err
		}
	}
	if ueIP != nil {
		t.routes = append(t.routes, defaultRoute(ueIP, t.table))
		return t.addRule(&net.IPNet{IP: ueIP, Mask: net.CIDRMask(32, 32)})
	}
	return nil
}

// routingTable returns the routing table of a UE, which needs no bookkeeping: from
// 0.128.0.0, the index of its val<MSIN>, unique on the host even when another
// process has a UE with the same address. A UE on a shared device has no val<MSIN>:
// its IPv4 address read as a number, from 1.0.0.0. Both are far above the tables of
// the kernel (0, 253 to 255) and those administrators number.
func routingTable(ueIP net.IP, endpoint netlink.Link) int {
	if endpoint != nil {
		return 1<<23 + endpoint.Attrs().Index
	}
	return int(binary.BigEndian.Uint32(ueIP))
}

// addRule sends what the UE's applications send from source to its routing table,
// unless its VRF does. A killed run may have left such a rule, to another table.
func (t *tunnel) addRule(source *net.IPNet) error {
	rule := netlink.NewRule()
	rule.Priority, rule.Src = 100, source
	for netlink.RuleDel(rule) == nil {
	}
	if t.vrf != nil {
		return nil
	}
	rule.Table = t.table
	if err := netlink.RuleAdd(rule); err != nil {
		return fmt.Errorf("routing policy: %w", err)
	}
	t.rules = append(t.rules, rule)
	return nil
}

// defaultRoute leads from the UE's routing table to the Data Network, for the IP
// version of source, through the device that setupTunnel gives it.
func defaultRoute(source net.IP, table int) *netlink.Route {
	route := &netlink.Route{
		Dst:      &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		Scope:    netlink.SCOPE_LINK,                                      // scope link
		Protocol: 4,                                                       // proto static
		Priority: 1,                                                       // metric 1
		Src:      source,
		Table:    table,
	}
	if source.To4() == nil {
		route.Dst = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
	}
	return route
}

// leave closes the datapath of the gNB the UE left, unless that was done already.
func (t *tunnel) leave(left *datapath) {
	if left != nil && t.left == left {
		(*left).close()
		t.left = nil
	}
}

// Release removes the tunnel from the host: its routes and rules first, so that
// nothing is routed to a datapath that is going away.
func (t *tunnel) Release() {
	for _, route := range t.routes {
		// ESRCH: it was never installed, or went with its device.
		if err := netlink.RouteDel(route); err != nil && !errors.Is(err, syscall.ESRCH) {
			log.Warn("[UE][GTP] Unable to remove route ", route, ": ", err)
		}
	}
	for _, rule := range t.rules {
		if err := netlink.RuleDel(rule); err != nil {
			log.Warn("[UE][GTP] Unable to remove ", rule, ": ", err)
		}
	}
	t.leave(t.left)
	if t.datapath != nil {
		t.datapath.close()
	}
	if t.endpoint != nil {
		removeLink(t.endpoint.Attrs().Name)
	}
	if t.vrf != nil {
		removeLink(t.vrf.Name)
	}
}

// removeLink removes the device name, if there is one.
func removeLink(name string) {
	if link, err := netlink.LinkByName(name); err == nil {
		if err := netlink.LinkDel(link); err != nil {
			log.Warn("[UE][GTP] Unable to remove ", name, ": ", err)
		}
	}
}

// logTunnel tells how to do traffic from the UE's address ueIp, IPv4 or IPv6.
func logTunnel(ue *context.UEContext, nameInf, ueIp string) {
	vrfInf := fmt.Sprintf("vrf%s", ue.GetMsin())
	log.Info(fmt.Sprintf("[UE][GTP] Interface %s has successfully been configured for UE %s", nameInf, ueIp))
	switch ue.TunnelMode {
	case config.TunnelTun, config.TunnelShared:
		log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE by binding to IP %s, eg:", ueIp))
		log.Info(fmt.Sprintf("[UE][GTP] iperf3 -B %s -c IPERF_SERVER -p PORT -t 9000", ueIp))
	case config.TunnelVrf:
		log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE using VRF %s, eg:", vrfInf))
		log.Info(fmt.Sprintf("[UE][GTP] sudo ip vrf exec %s iperf3 -c IPERF_SERVER -p PORT -t 9000", vrfInf))
	}
}
