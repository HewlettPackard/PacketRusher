/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"fmt"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"

	gtpTunnel "github.com/free5gc/go-gtp5gnl/tuncmd"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"syscall"
)

// addPDR and addFAR prefer the shared device's long-lived netlink client and fall
// back to the per-call wrapper when a UE owns its device outright.
func addPDR(link sharedGTPDevice, args []string) error {
	if link != nil {
		return link.AddPDR(args)
	}

	return addTunnelPDR(args)
}

func addFAR(link sharedGTPDevice, args []string) error {
	if link != nil {
		return link.AddFAR(args)
	}

	return addTunnelFAR(args)
}

// addPDRVerified installs a PDR and, when PR_VERIFY_RULES=1, confirms it reached the
// datapath, retrying once. A silent install failure is indistinguishable from success at
// setup time -- the UE keeps its address, rule and route and simply never passes traffic
// -- so the only way to catch it is to read the rule back. It returns an error when the
// PDR is not known to be installed, so the caller treats it as a failed create.
func addPDRVerified(link sharedGTPDevice, id uint32, args []string, label string) error {
	err := addPDR(link, args)
	if link == nil || !verifyRules() {
		return err
	}

	installed, checked := link.PDRInstalled(id)
	if !checked {
		// Nothing to read the rule back with: the create's own result is all there is.
		return err
	}
	if installed {
		return nil
	}

	log.Warn("[UE][GTP] ", label, " PDR ", id, " did not reach the datapath (", err, "), retrying")

	if err := addPDR(link, args); err != nil {
		return fmt.Errorf("retry of %s PDR %d: %w", label, id, err)
	}

	if installed, checked := link.PDRInstalled(id); checked && !installed {
		return fmt.Errorf("%s PDR %d still absent after retry; this UE will not pass traffic", label, id)
	}

	return nil
}

// sharedDevice is what a UE uses of the gNB's shared device to give back its tunnel.
type sharedDevice interface {
	RemoveAddress(ueIP string)
	Release(ids gtp.RuleIDs)
}

type sharedGTPDevice interface {
	sharedDevice
	Name() string
	Take() (gtp.RuleIDs, error)
	AddPDR([]string) error
	AddFAR([]string) error
	AddQER([]string) error
	EnsureQER(int64, func(uint32) error) (uint32, bool)
	PDRInstalled(uint32) (bool, bool)
}

// Seams for the tests: removing a policy rule or route needs privileges they lack.
var (
	ruleDel            = netlink.RuleDel
	routeDel           = netlink.RouteDel
	addTunnelPDR       = gtpTunnel.CmdAddPDR
	addTunnelFAR       = gtpTunnel.CmdAddFAR
	addTunnelQER       = gtpTunnel.CmdAddQER
	addTunnelAddress   = netlink.AddrAdd
	addTunnelRule      = netlink.RuleAdd
	addTunnelVRF       = netlink.LinkAdd
	setTunnelMaster    = netlink.LinkSetMaster
	setTunnelUp        = netlink.LinkSetUp
	replaceTunnelRoute = netlink.RouteReplace
	setTunnelMTU       = gtp.SetTunnelMTU
	setUEEndpointMTU   = netlink.LinkSetMTU
	makeUEEndpoint     = createUEEndpoint
	sharedDeviceFor    = func(msg gnbContext.UEMessage) sharedGTPDevice {
		if msg.GtpDevice == nil {
			return nil
		}
		return msg.GtpDevice
	}
)

// sharedTunnel is what one UE's session holds on a device it shares with the other
// UEs of its gNB, recorded as setup installs it so that releasing the session
// removes exactly these and nothing of the device's or of another UE's.
type sharedTunnel struct {
	dev         sharedDevice
	ueIP        string
	ids         gtp.RuleIDs
	rule        *netlink.Rule
	route       *netlink.Route
	keepAddress bool
	table       *routingTableReservation
}

// release removes the UE's route first, so nothing is routed to rules that are going
// away, then its policy rule, address and GTP-U rules, and clears the session's hold on
// them so nothing else removes them again.
func (t *sharedTunnel) release(pduSession *context.UEPDUSession) {
	routingRemoved := true
	if t.route != nil {
		routingRemoved = routingObjectRemoved(routeDel(t.route))
	}
	if t.rule != nil {
		routingRemoved = routingObjectRemoved(ruleDel(t.rule)) && routingRemoved
	}
	if !t.keepAddress {
		t.dev.RemoveAddress(t.ueIP)
	}
	t.dev.Release(t.ids)
	if routingRemoved {
		t.table.release()
	} else {
		t.table.quarantine()
		log.Warn("[UE][GTP] Keeping routing table reservation after incomplete cleanup for ", t.ueIP)
	}

	if pduSession != nil {
		pduSession.SetTunRoute(nil)
		pduSession.SetTunRule(nil)
		pduSession.SetTunInterface(nil)
	}
}

func SetupGtpInterface(ue *context.UEContext, msg gnbContext.UEMessage) {
	gnbPduSession := msg.GNBPduSessions[0]
	pduSession, err := ue.GetPduSession(uint8(gnbPduSession.GetPduSessionId()))
	if pduSession == nil || err != nil {
		log.Error("[GNB][GTP] Aborting the setup of PDU Session ", gnbPduSession.GetPduSessionId(), ", this PDU session was not succesfully configured on the UE's side.")
		return
	}
	previousLink := pduSession.GetTunInterface()
	previousRule := pduSession.GetTunRule()
	previousGnbIP := pduSession.GetGnbIp()
	previousGNBPDU := pduSession.GnbPduSession
	committed := false
	defer func() {
		if !committed {
			pduSession.GnbPduSession = previousGNBPDU
			pduSession.SetGnbIp(previousGnbIP)
		}
	}()
	pduSession.GnbPduSession = gnbPduSession

	if ue.TunnelMode == config.TunnelDisabled {
		committed = true
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

	// A repeated setup on the same N3 endpoint updates its existing rules. A
	// second socket cannot bind the same address while the first remains active.
	if previousLink != nil && ue.TunnelMode != config.TunnelShared && pduSession.GetGnbIp() == msg.GnbIp {
		if err := pduSession.UpdateTunnel(gnbPduSession, msg.GnbIp); err != nil {
			log.Error("[UE][GTP] Unable to update existing tunnel: ", err)
			if errors.Is(err, errTunnelRollback) {
				pduSession.ReleaseTunnel()
				log.Error("[UE][GTP] Released tunnel after unsuccessful rule rollback")
			}
		} else {
			committed = true
		}
		return
	}

	// get UE GNB IP.
	pduSession.SetGnbIp(msg.GnbIp)

	ueGnbIp := pduSession.GetGnbIp()
	upfIp := pduSession.GnbPduSession.GetUpfIp()
	qfi := pduSession.GnbPduSession.GetQosId()
	ueIp := pduSession.GetIp()
	msin := ue.GetMsin()
	vrfInf := fmt.Sprintf("vrf%s", msin)

	// In shared mode every UE of this gNB rides the gNB's device and contributes its
	// own rules, so the identifiers have to be allocated rather than assumed.
	var (
		nameInf   string
		ids       gtp.RuleIDs
		sharedFor sharedGTPDevice
		held      *sharedTunnel
		setupDone bool
		dedicated *dedicatedTunnel
	)

	// Failed staging returns only the newly acquired resources. The previous
	// tunnel remains usable, and one UE's failure does not stop the load test.
	failed := func(args ...any) {
		if sharedFor != nil {
			log.Error(append(args, "; no tunnel for UE ", ueIp)...)
		} else {
			log.Error(args...)
		}
	}

	if ue.TunnelMode == config.TunnelShared {

		dev := sharedDeviceFor(msg)
		if dev == nil {
			log.Error("[GNB][GTP] gNB has no shared GTP-U device; no tunnel for UE ", msin)
			return
		}

		ids, err = dev.Take()
		if err != nil {
			log.Error("[GNB][GTP] No tunnel for UE ", msin, " on ", dev.Name(), ": ", err)
			return
		}

		nameInf = dev.Name()
		sharedFor = dev
		held = &sharedTunnel{dev: dev, ueIP: ueIp, ids: ids, keepAddress: true}

		defer func() {
			if !setupDone {
				held.release(nil)
			}
		}()
	} else {
		// val<MSIN> owns the UE address for the whole PDU session. The socket-bound
		// GTP backends alternate so the target can be installed before retiring source.
		nameInf = fmt.Sprintf("gtp0%s", msin)
		if previousLink != nil && previousLink.Attrs().Name == nameInf {
			nameInf = fmt.Sprintf("gtp1%s", msin)
		}
		ids = gtp.DedicatedRuleIDs()
		dedicated, err = startDedicatedTunnel(nameInf, ueGnbIp)
		if err != nil {
			failed("[GNB][GTP] Unable to create Kernel GTP interface: ", err)
			return
		}
		defer func() {
			if !setupDone {
				dedicated.release(false)
			}
		}()
	}

	// Create FAR for downlink: forward decapsulated packets towards the UE.
	cmdAddFar := []string{nameInf,
		strconv.FormatUint(uint64(ids.FARDown), 10), // FAR ID
		"--action", "2", // Apply Action = FORW
	}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := addFAR(sharedFor, cmdAddFar); err != nil {
		failed("[GNB][GTP] Unable to create FAR: ", err)
		return
	}

	// Create FAR for uplink: encapsulate towards the UPF.
	cmdAddFar = []string{nameInf,
		strconv.FormatUint(uint64(ids.FARUp), 10), // FAR ID
		"--action", "2", // Apply Action = FORW
		"--hdr-creation", "0", strconv.FormatUint(uint64(gnbPduSession.GetTeidUplink()), 10), upfIp, "2152", // Outer Header Creation
	}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := addFAR(sharedFor, cmdAddFar); err != nil {
		failed("[UE][GTP] Unable to create FAR ", err)
		return
	}

	// Create PDR for downlink: GTP-U from the UPF on this gNB's F-TEID.
	cmdAddPdr := []string{nameInf,
		strconv.FormatUint(uint64(ids.PDRDown), 10), // PDR ID
		"--pcd", "1", // Precedence = 1
		"--hdr-rm", "0", // Outer Header Removal = GTP-U/UDP/IPv4
		"--ue-ipv4", ueIp, // UE IP Address
		"--f-teid", strconv.FormatUint(uint64(gnbPduSession.GetTeidDownlink()), 10), msg.GnbIp.String(), // F-TEID
		"--far-id", strconv.FormatUint(uint64(ids.FARDown), 10), // FAR ID
		"--src-intf", "1", // Source Interface = Core
	}
	log.Debug("[UE][GTP] Setting up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))
	if err := addPDRVerified(sharedFor, ids.PDRDown, cmdAddPdr, "downlink"); err != nil {
		failed("[GNB][GTP] Unable to create downlink PDR: ", err)
		return
	}

	// Create PDR for uplink: packets from the UE's address.
	cmdAddPdr = []string{nameInf,
		strconv.FormatUint(uint64(ids.PDRUp), 10), // PDR ID
		"--pcd", "2", // Precedence = 2
		"--ue-ipv4", ueIp, // UE IP Address
		"--far-id", strconv.FormatUint(uint64(ids.FARUp), 10), // FAR ID
		"--src-intf", "0", // Source Interface = Access
		"--gtpu-src-ip", ueGnbIp.String(), // GTP-U source IP address (not part of PFCP spec)
	}
	if qfi > 0 {
		cmdAddQer := func(id uint32) []string {
			return []string{nameInf,
				strconv.FormatUint(uint64(id), 10),  // QER ID
				"--qfi", strconv.FormatInt(qfi, 10), // QFI
			}
		}

		// On a shared device QERs belong to the device, one per QFI, not to the UE:
		// gtp5g refuses a QER that already exists, so creating it per UE failed for
		// every UE after the first -- and because that error used to abandon the
		// setup, those UEs silently got no address, no rule and no route while still
		// reporting a session. Keying by QFI keeps a UE given another QFI from being
		// marked with the first UE's.
		qerID, qerUsable := uint32(gtp.DedicatedQERID), true
		if sharedFor != nil {
			qerID, qerUsable = sharedFor.EnsureQER(qfi, func(id uint32) error {
				log.Debug("[UE][GTP] Setting Up QFI ", strings.Join(cmdAddQer(id), " "))
				return sharedFor.AddQER(cmdAddQer(id))
			})
		} else {
			log.Debug("[UE][GTP] Setting Up QFI", strings.Join(cmdAddQer(qerID), " "))
			if err := addTunnelQER(cmdAddQer(qerID)); err != nil {
				failed("[UE][GTP] Unable to create QER:", err)
				return
			}
		}

		if qerUsable {
			cmdAddPdr = append(cmdAddPdr,
				"--qer-id", strconv.FormatUint(uint64(qerID), 10), // QER ID
			)
		}
	}

	log.Debug("[UE][GTP] Setting Up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))
	if err := addPDRVerified(sharedFor, ids.PDRUp, cmdAddPdr, "uplink"); err != nil {
		failed("[UE][GTP] Unable to create uplink PDR: ", err)
		return
	}

	// The shared device stays gNB-owned; dedicated mode keeps its UE-facing
	// address on a stable dummy while the GTP socket/backend changes.
	link, err := findTunnelLink(nameInf)
	if err != nil {
		failed("[UE][GTP] Tunnel link unavailable: ", err)
		return
	}
	addressLink := link
	if dedicated != nil {
		addressLink = pduSession.GetUEInterface()
		if addressLink == nil {
			addressLink, err = makeUEEndpoint(fmt.Sprintf("val%s", msin))
			if err != nil {
				failed("[UE][GTP] UE endpoint unavailable: ", err)
				return
			}
			dedicated.endpoint = addressLink
			dedicated.endpointOwned = true
		}
	}
	if dedicated != nil {
		if err := setTunnelMTU(link, ueGnbIp, ue.TunnelMTU); err != nil {
			failed("[UE][GTP] Unable to configure tunnel MTU: ", err)
			return
		}
	}
	addrTun := &netlink.Addr{IPNet: &net.IPNet{IP: net.ParseIP(ueIp).To4(), Mask: net.CIDRMask(32, 32)}}
	sameAddressLink := previousLink != nil && previousLink.Attrs().Index == addressLink.Attrs().Index
	if dedicated != nil && pduSession.GetUEInterface() != nil {
		sameAddressLink = true
	}
	if err := addTunnelAddress(addressLink, addrTun); err != nil && !(sameAddressLink && errors.Is(err, syscall.EEXIST)) {
		failed("[UE][DATA] Error adding UE address: ", err)
		return
	}
	if held != nil {
		held.keepAddress = sameAddressLink
	}
	if dedicated != nil {
		dedicated.endpoint = addressLink
	}

	// The session keeps its table across handover, even if the old UPF TEID is
	// recycled for another UE. Staging owns only a newly acquired reservation.
	table, tableOwned, err := sessionRoutingTables.reserve(pduSession)
	if err != nil {
		failed("[UE][DATA] Unable to reserve routing table: ", err)
		return
	}
	tableId := table.table
	if tableOwned {
		if held != nil {
			held.table = table
		} else {
			dedicated.table = table
		}
	}
	// Configure routing policy or VRF for the UE.
	switch ue.TunnelMode {
	case config.TunnelTun, config.TunnelShared:
		rule := previousRule
		if rule == nil {
			rule = netlink.NewRule()
			rule.Priority = 100
			rule.Table = int(tableId)
			rule.Src = addrTun.IPNet
			if err := addTunnelRule(rule); err != nil {
				failed("[UE][DATA] Unable to create routing policy: ", err)
				return
			}
		}
		if held != nil && previousRule == nil {
			held.rule = rule
		}
		if dedicated != nil && previousRule == nil {
			dedicated.rule = rule
		}

	case config.TunnelVrf:
		vrfDevice := pduSession.GetVrfDevice()
		if vrfDevice == nil {
			vrfDevice = &netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: vrfInf}, Table: tableId}
			if err := addTunnelVRF(vrfDevice); err != nil {
				failed("[UE][DATA] Unable to create VRF: ", err)
				return
			}
			dedicated.vrf = vrfDevice
			dedicated.vrfOwned = true
		}
		if err := setTunnelMaster(link, vrfDevice); err != nil {
			failed("[UE][DATA] Unable to attach GTP backend to VRF: ", err)
			return
		}
		if err := setTunnelMaster(addressLink, vrfDevice); err != nil {
			failed("[UE][DATA] Unable to attach UE address to VRF: ", err)
			return
		}
		if err := setTunnelUp(vrfDevice); err != nil {
			failed("[UE][DATA] Unable to enable VRF: ", err)
			return
		}
		dedicated.vrf = vrfDevice

	}

	// Insert default route from the UE to the Data Network.
	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: link.Attrs().Index,                                      // dev val<MSIN>
		Scope:     netlink.SCOPE_LINK,                                      // scope link
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Src:       net.ParseIP(ueIp).To4(),
		Table:     int(tableId), // table <ECI>
	}
	if dedicated != nil {
		err = replaceRouteWithEndpointMTU(route, addressLink, link.Attrs().MTU, !dedicated.endpointOwned)
	} else {
		err = replaceTunnelRoute(route)
	}
	if err != nil {
		failed("[GNB][GTP] Unable to create Kernel Route ", err)
		return
	}
	// Commit after the target route is installed. The previous cleanup runs in
	// retirement mode, so it cannot remove the policy rule or the replaced route.
	pduSession.SetTunInterface(link)
	pduSession.SetTunRoute(route)
	if previousRule != nil {
		pduSession.SetTunRule(previousRule)
	}
	if held != nil {
		held.table = table
		held.route = route
		if held.rule == nil {
			held.rule = previousRule
		}
		pduSession.SetTunRule(held.rule)
		held.keepAddress = false // This completed binding now owns the address.
		pduSession.ReplaceTunnelCleanup(func(retiring bool) {
			if retiring {
				held.retireOn(link, pduSession.GetTunInterface())
			}
			held.release(nil)
		})
	} else {
		dedicated.table = table
		dedicated.route = route
		dedicated.endpointOwned = true
		dedicated.vrfOwned = dedicated.vrf != nil
		if dedicated.rule == nil {
			dedicated.rule = previousRule
		}
		pduSession.SetTunRule(dedicated.rule)
		pduSession.SetUEInterface(addressLink)
		pduSession.SetVrfDevice(dedicated.vrf)
		dedicated.activeRules = dedicatedRuleSet(nameInf, gnbPduSession, ueIp, msg.GnbIp)
		pduSession.ReplaceTunnelCleanup(dedicated.release)
		pduSession.SetTunnelUpdate(func(pdu *gnbContext.GnbPDUSession, ip netip.Addr) error { return dedicated.refresh(pdu, ueIp, ip) })
	}
	setupDone = true
	committed = true

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

// A retained UE endpoint still belongs to the source until the new route commits.
// Match the target backend's already validated MTU, and restore the source MTU
// if either the endpoint update or route replacement fails.
func replaceRouteWithEndpointMTU(route *netlink.Route, endpoint netlink.Link, mtu int, preserve bool) error {
	previous := endpoint.Attrs().MTU
	err := setUEEndpointMTU(endpoint, mtu)
	if err == nil {
		endpoint.Attrs().MTU = mtu
		err = replaceTunnelRoute(route)
	}
	if err != nil && preserve {
		if restoreErr := setUEEndpointMTU(endpoint, previous); restoreErr != nil {
			return errors.Join(err, fmt.Errorf("source endpoint MTU rollback failed: %w", restoreErr))
		}
		endpoint.Attrs().MTU = previous
	}
	return err
}

// A replaced route and its policy rule belong to the new binding. On the same
// device, so does the address; on another gNB, remove only the source's copy.
func (t *sharedTunnel) retireOn(source, target netlink.Link) {
	t.route, t.rule = nil, nil
	t.table = nil
	t.keepAddress = target != nil && source.Attrs().Index == target.Attrs().Index
}
