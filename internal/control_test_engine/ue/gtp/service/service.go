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

	gtpLink "github.com/free5gc/go-gtp5gnl/linkcmd"
	gtpTunnel "github.com/free5gc/go-gtp5gnl/tuncmd"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"net"
	"strconv"
	"strings"
	"time"
)

// addPDR and addFAR prefer the shared device's long-lived netlink client and fall
// back to the per-call wrapper when a UE owns its device outright.
func addPDR(link *gtp.Device, args []string) error {
	if link != nil {
		return link.AddPDR(args)
	}

	return gtpTunnel.CmdAddPDR(args)
}

func addFAR(link *gtp.Device, args []string) error {
	if link != nil {
		return link.AddFAR(args)
	}

	return gtpTunnel.CmdAddFAR(args)
}

// addPDRVerified installs a PDR and, when PR_VERIFY_RULES=1, confirms it reached the
// datapath, retrying once. A silent install failure is indistinguishable from success at
// setup time -- the UE keeps its address, rule and route and simply never passes traffic
// -- so the only way to catch it is to read the rule back. It returns an error when the
// PDR is not known to be installed, so the caller stops as it does for a failed FAR.
func addPDRVerified(link *gtp.Device, id uint32, args []string, label string) error {
	err := addPDR(link, args)
	if link == nil || !verifyRules() {
		if err != nil {
			log.Error("[UE][GTP] Unable to create ", label, " PDR: ", err)
		}
		return err
	}

	if link.PDRInstalled(id) {
		return nil
	}

	log.Warn("[UE][GTP] ", label, " PDR ", id, " did not reach the datapath (", err, "), retrying")

	if err := addPDR(link, args); err != nil {
		log.Error("[UE][GTP] Retry of ", label, " PDR ", id, " failed: ", err)
		return err
	}

	if !link.PDRInstalled(id) {
		log.Error("[UE][GTP] ", label, " PDR ", id, " still absent after retry; this UE will not pass traffic")
		return fmt.Errorf("%s PDR %d absent after retry", label, id)
	}

	return nil
}

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
		sharedFor *gtp.Device
		setupDone bool

		// What this setup installs outside the device, recorded as it goes so that
		// releasing the session removes exactly these.
		tunRule  *netlink.Rule
		tunRoute *netlink.Route
	)

	if ue.TunnelMode == config.TunnelShared {
		// A handover or a repeated setup sets the tunnel up again for a session that
		// already has one, possibly on another gNB's device. Give back what it holds
		// first -- its route, policy rule, address, rules and identifiers. Otherwise
		// they stay behind: adding the address again on the same device fails, and
		// if the uplink TEID changed, the old policy rule still matches this UE's
		// address and sends it to the old table.
		pduSession.ReleaseTunnel()

		dev := msg.GtpDevice
		if dev == nil {
			log.Error("[GNB][GTP] gNB has no shared GTP-U device; no tunnel for UE ", msin)
			return
		}

		nameInf = dev.Name()
		ids = dev.Take()
		sharedFor = dev

		// Until setup completes, a failure hands everything back at once; afterwards
		// the UE does, on Terminate or on the next setup of this session. Only this
		// UE's parts come off: the device is the gNB's.
		slot := ids
		release := func() {
			if tunRoute != nil {
				_ = netlink.RouteDel(tunRoute)
			}
			if tunRule != nil {
				_ = netlink.RuleDel(tunRule)
			}
			dev.RemoveAddress(ueIp)
			dev.Release(slot)

			// The session no longer holds these, so nothing else may remove them.
			pduSession.SetTunRoute(nil)
			pduSession.SetTunRule(nil)
			pduSession.SetTunInterface(nil)
		}
		defer func() {
			if setupDone {
				pduSession.SetTunnelRelease(release)
			} else {
				release()
			}
		}()
	} else {
		nameInf = fmt.Sprintf("val%s", msin)
		ids = gtp.DedicatedRuleIDs()
		stopSignal := make(chan bool)

		_ = gtpLink.CmdDel(nameInf)

		if pduSession.GetStopSignal() != nil {
			close(pduSession.GetStopSignal())
			time.Sleep(time.Second)
		}

		go func() {
			// This function should not return as long as the GTP-U UDP socket is open
			if err := gtpLink.CmdAddWithStopCh(nameInf, 1, 131072, ueGnbIp.String(), "", stopSignal); err != nil {
				log.Fatal("[GNB][GTP] Unable to create Kernel GTP interface: ", err, msin, nameInf)
				return
			}
		}()

		pduSession.SetStopSignal(stopSignal)

		time.Sleep(time.Second)
	}

	// Create FAR for downlink: forward decapsulated packets towards the UE.
	cmdAddFar := []string{nameInf,
		strconv.FormatUint(uint64(ids.FARDown), 10), // FAR ID
		"--action", "2", // Apply Action = FORW
	}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := addFAR(sharedFor, cmdAddFar); err != nil {
		log.Error("[GNB][GTP] Unable to create FAR: ", err)
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
		log.Error("[UE][GTP] Unable to create FAR ", err)
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
			log.Debug("[UE][GTP] Setting Up QFI ", strings.Join(cmdAddQer(qerID), " "))
			if err := gtpTunnel.CmdAddQER(cmdAddQer(qerID)); err != nil {
				log.Error("[UE][GTP] Unable to create QER: ", err)

				qerUsable = false
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
		return
	}

	// Find TUN network interface.
	link, _ := netlink.LinkByName(nameInf)
	pduSession.SetTunInterface(link)

	// Add UE IP Address onto the TUN network interface.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ueIp).To4(),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
	}
	if err := netlink.AddrAdd(link, addrTun); err != nil {
		log.Error("[UE][DATA] Error in adding IP for virtual interface: ", err)
		return
	}

	// Configure routing policy or VRF for the UE.
	//
	// The uplink TEID identifies the table, but it cannot be used raw. Linux reserves
	// table 255 (local), 254 (main) and 253 (default), and the local table is consulted
	// before every other rule -- so a UE whose TEID happens to be 255 installs a
	// default route that captures the host's own traffic, including the simulator's
	// SCTP association to the AMF. TEIDs restart from 1 whenever the SMF restarts, so
	// at a few hundred UEs this is not a corner case: it broke a 10 500-UE run at UE
	// 253 and left the host routing through a GTP device.
	tableId := routeTableOffset + gnbPduSession.GetTeidUplink()
	switch ue.TunnelMode {
	case config.TunnelTun, config.TunnelShared:
		rule := netlink.NewRule()
		rule.Priority = 100
		rule.Table = int(tableId)
		rule.Src = addrTun.IPNet
		_ = netlink.RuleDel(rule)

		if err := netlink.RuleAdd(rule); err != nil {
			log.Error("[UE][DATA] Unable to create routing policy rule for UE: ", err)
			return
		}
		tunRule = rule
		pduSession.SetTunRule(rule)
	case config.TunnelVrf:
		vrfDevice := &netlink.Vrf{
			LinkAttrs: netlink.LinkAttrs{
				Name: vrfInf,
			},
			Table: tableId,
		}
		_ = netlink.LinkDel(vrfDevice)

		if err := netlink.LinkAdd(vrfDevice); err != nil {
			log.Fatal("[UE][DATA] Unable to create VRF for UE", err)
			return
		}

		if err := netlink.LinkSetMaster(link, vrfDevice); err != nil {
			log.Fatal("[UE][DATA] Unable to set GTP tunnel as slave of VRF interface", err)
			return
		}

		if err := netlink.LinkSetUp(vrfDevice); err != nil {
			log.Fatal("[UE][DATA] Unable to set interface VRF UP", err)
			return
		}
		pduSession.SetVrfDevice(vrfDevice)
	}

	// Insert default route from the UE to the Data Network.
	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: link.Attrs().Index,                                      // dev val<MSIN>
		Scope:     netlink.SCOPE_LINK,                                      // scope link
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Table:     int(tableId),                                            // table <ECI>
	}
	// Without its route the UE has an address and GTP-U rules but no way to reach
	// them, so it is not set up: stop here, and in shared mode hand everything back.
	if err := netlink.RouteReplace(route); err != nil {
		log.Error("[GNB][GTP] Unable to create Kernel Route: ", err, "; no tunnel for UE ", ueIp)
		return
	}
	tunRoute = route
	pduSession.SetTunRoute(route)
	setupDone = true

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
