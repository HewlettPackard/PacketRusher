/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"fmt"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
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
	nameInf := fmt.Sprintf("val%s", msin)
	vrfInf := fmt.Sprintf("vrf%s", msin)
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

	// Create FAR for uplink.
	cmdAddFar := []string{nameInf,
		"1",             // FAR ID = 1
		"--action", "2", // Apply Action = FORW
	}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
		log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
		return
	}

	// Create FAR for downlink.
	cmdAddFar = []string{nameInf,
		"2",             // FAR ID = 2
		"--action", "2", // Apply Action = FORW
		"--hdr-creation", "0", strconv.FormatUint(uint64(gnbPduSession.GetTeidUplink()), 10), upfIp, "2152", // Outer Header Creation
	}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
		log.Fatal("[UE][GTP] Unable to create FAR ", err)
		return
	}

	// Create PDR for uplink.
	cmdAddPdr := []string{nameInf,
		"1",          // PDR ID = 1
		"--pcd", "1", // Precedence = 1
		"--hdr-rm", "0", // Outer Header Removal = GTP-U/UDP/IPv4
		"--ue-ipv4", ueIp, // UE IP Address
		"--f-teid", strconv.FormatUint(uint64(gnbPduSession.GetTeidDownlink()), 10), msg.GnbIp.String(), // F-TEID
		"--far-id", "1", // FAR ID = 1
		"--src-intf", "1", // Source Interface = Core
	}
	log.Debug("[UE][GTP] Setting up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))
	if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
		log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
		return
	}

	cmdAddPdr = []string{nameInf,
		"2",          // PDR ID = 2
		"--pcd", "2", // Precedence = 2
		"--ue-ipv4", ueIp, // UE IP Address
		"--far-id", "2", // FAR ID = 2
		"--src-intf", "0", // Source Interface = Access
		"--gtpu-src-ip", ueGnbIp.String(), // GTP-U source IP address (not part of PFCP spec)
	}
	if qfi > 0 {
		// Create QER for downlink.
		cmdAddQer := []string{nameInf,
			"1",                                 // QER ID = 1
			"--qfi", strconv.FormatInt(qfi, 10), // QFI
		}
		log.Debug("[UE][GTP] Setting Up QFI", strings.Join(cmdAddQer, " "))
		if err := gtpTunnel.CmdAddQER(cmdAddQer); err != nil {
			log.Fatal("[UE][GTP] Unable to create QER:", err)
			return
		}
		cmdAddPdr = append(cmdAddPdr,
			"--qer-id", "1", // QER ID = 1
		)
	}

	// Create PDR for downlink.
	log.Debug("[UE][GTP] Setting Up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))
	if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
		log.Fatal("[UE][GTP] Unable to create FAR ", err)
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
		log.Fatal("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}

	// Configure routing policy or VRF for the UE.
	tableId := gnbPduSession.GetTeidUplink()
	switch ue.TunnelMode {
	case config.TunnelTun:
		rule := netlink.NewRule()
		rule.Priority = 100
		rule.Table = int(tableId)
		rule.Src = addrTun.IPNet
		_ = netlink.RuleDel(rule)

		if err := netlink.RuleAdd(rule); err != nil {
			log.Fatal("[UE][DATA] Unable to create routing policy rule for UE", err)
			return
		}
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
	if err := netlink.RouteReplace(route); err != nil {
		log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}
	pduSession.SetTunRoute(route)

	log.Info(fmt.Sprintf("[UE][GTP] Interface %s has successfully been configured for UE %s", nameInf, ueIp))
	switch ue.TunnelMode {
	case config.TunnelTun:
		log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE by binding to IP %s, eg:", ueIp))
		log.Info(fmt.Sprintf("[UE][GTP] iperf3 -B %s -c IPERF_SERVER -p PORT -t 9000", ueIp))
	case config.TunnelVrf:
		log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE using VRF %s, eg:", vrfInf))
		log.Info(fmt.Sprintf("[UE][GTP] sudo ip vrf exec %s iperf3 -c IPERF_SERVER -p PORT -t 9000", vrfInf))
	}
}
