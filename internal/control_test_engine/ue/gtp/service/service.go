/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"fmt"
	gtpLink "my5G-RANTester/internal/cmd/gogtp5g-link"
	gtpTunnel "my5G-RANTester/internal/cmd/gogtp5g-tunnel"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"

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
	pduSession.GnbPduSession = gnbPduSession
	if err != nil {
		log.Error("[GNB][GTP] ", err)
		return
	}

	if !ue.IsTunnelEnabled() {
		log.Info(fmt.Sprintf("[UE][GTP] Interface for UE %s has not been created. Tunnel has been disabled.", ue.GetMsin()))
		return
	}

	if pduSession.Id != 1 {
		log.Warn("[GNB][GTP] Only one tunnel per UE is supported for now, no tunnel will be created for second PDU Session of given UE")
		return
	}

	// get UE GNB IP.
	pduSession.SetGnbIp(net.ParseIP(msg.GnbIp))

	ueGnbIp := pduSession.GetGnbIp()
	upfIp := pduSession.GnbPduSession.GetUpfIp()
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
		if err := gtpLink.CmdAdd(nameInf, 1, ueGnbIp.String(), stopSignal); err != nil {
			log.Fatal("[GNB][GTP] Unable to create Kernel GTP interface: ", err, msin, nameInf)
			return
		}
	}()

	pduSession.SetStopSignal(stopSignal)

	time.Sleep(time.Second)

	cmdAddFar := []string{nameInf, "1", "--action", "2"}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
		log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
		return
	}

	cmdAddFar = []string{nameInf, "2", "--action", "2", "--hdr-creation", "0", fmt.Sprint(gnbPduSession.GetTeidUplink()), upfIp, "2152"}
	log.Debug("[UE][GTP] Setting up GTP Forwarding Action Rule for ", strings.Join(cmdAddFar, " "))
	if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
		log.Fatal("[UE][GTP] Unable to create FAR ", err)
		return
	}

	cmdAddPdr := []string{nameInf, "1", "--pcd", "1", "--hdr-rm", "0", "--ue-ipv4", ueIp, "--f-teid", fmt.Sprint(gnbPduSession.GetTeidDownlink()), msg.GnbIp, "--far-id", "1"}
	log.Debug("[UE][GTP] Setting up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))

	if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
		log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
		return
	}

	cmdAddPdr = []string{nameInf, "2", "--pcd", "2", "--ue-ipv4", ueIp, "--far-id", "2"}
	log.Debug("[UE][GTP] Setting Up GTP Packet Detection Rule for ", strings.Join(cmdAddPdr, " "))
	if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
		log.Fatal("[UE][GTP] Unable to create FAR ", err)
		return
	}

	netUeIp := net.ParseIP(ueIp)
	// add an IP address to a link device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   netUeIp.To4(),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
	}

	link, _ := netlink.LinkByName(nameInf)
	pduSession.SetTunInterface(link)

	if err := netlink.AddrAdd(link, addrTun); err != nil {
		log.Fatal("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}

	tableId, _ := strconv.Atoi(fmt.Sprint(gnbPduSession.GetTeidUplink()))

	vrfDevice := &netlink.Vrf{
		LinkAttrs: netlink.LinkAttrs{
			Name: vrfInf,
		},
		Table: uint32(tableId),
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

	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: link.Attrs().Index,                                      // dev gtp-<ECI>
		Scope:     netlink.SCOPE_LINK,                                      // scope link
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Table:     tableId,                                                 // table <ECI>
	}

	if err := netlink.RouteReplace(route); err != nil {
		log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}
	pduSession.SetTunRoute(route)

	log.Info(fmt.Sprintf("[UE][GTP] Interface %s has successfully been configured for UE %s", nameInf, ueIp))
	log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE using VRF %s, eg:", vrfInf))
	log.Info(fmt.Sprintf("[UE][GTP] sudo ip vrf exec %s iperf3 -c IPERF_SERVER -p PORT -t 9000", vrfInf))
}
