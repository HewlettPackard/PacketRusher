/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"encoding/binary"
	"fmt"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/lib/eupf/ebpf"
	"net"
	"os/exec"
	"strconv"
	"time"
	"unsafe"

	"github.com/cilium/ebpf/link"
	"github.com/lorenzosaino/go-sysctl"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
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

	if err := ebpf.IncreaseResourceLimits(); err != nil {
		log.Fatalf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := ebpf.NewBpfObjects()
	if err := bpfObjects.Load(); err != nil {
		log.Fatalf("Loading bpf objects failed: %s", err.Error())
	}

	/*if err := bpfObjects.ResizeAllMaps(1024, 1024, 1024); err != nil {
		log.Fatalf("Resizing bpf objects failed: %s", err.Error())
	}*/

	if err := sysctl.Set("net.ipv4.ip_forward", "1"); err != nil {
		log.Fatalf("Unable to enable IP forwarding: %s", err.Error())
	}

	tableId, _ := strconv.ParseInt(fmt.Sprint(gnbPduSession.GetTeidUplink()), 10, 32)
	ueVrfInf := fmt.Sprintf("vrf%s", ue.GetMsin())
	ueIp := pduSession.GetIp()

	/*
		ueTunInf := fmt.Sprintf("ue%s", ue.GetMsin())
		ueTunDevice := &netlink.Tuntap{
			LinkAttrs: netlink.LinkAttrs{
				Name: ueTunInf,
				MTU: 1400,
			},
			Mode: netlink.TUNTAP_MODE_TAP,
		}

		if err := netlink.LinkDel(ueTunDevice); err != nil {
			log.Warn("[UE][DATA] Unable to delete tun for UE", err)
		}

		if err := netlink.LinkAdd(ueTunDevice); err != nil {
			log.Fatal("[UE][DATA] Unable to create tun for UE", err)
			return
		}
		ueTunDevice2, _ := netlink.LinkByName(ueTunInf)

		pduSession.SetTunInterface(ueTunDevice2)*/

	netUeIp := net.ParseIP(ueIp).To4()
	upfIp := net.ParseIP(pduSession.GnbPduSession.GetUpfIp()).To4()
	gnbIp := net.ParseIP(msg.GnbIp).To4()

	ueVrfDevice := &netlink.Vrf{
		LinkAttrs: netlink.LinkAttrs{
			Name: ueVrfInf,
			MTU:  1400,
		},
		Table: uint32(tableId),
	}
	_ = netlink.LinkDel(ueVrfDevice)

	if err := netlink.LinkAdd(ueVrfDevice); err != nil {
		log.Fatal("[UE][DATA] Unable to create VRF for UE", err)
		return
	}

	// add an IP address to a n3Inf device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   netUeIp.To4(),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
	}

	if err := netlink.AddrAdd(ueVrfDevice, addrTun); err != nil {
		log.Fatal("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}

	/*
		if err := netlink.LinkSetUp(ueTunDevice2); err != nil {
			log.Fatal("[UE][DATA] Unable to set interface vEth UP", err)
			return
		}

		if err := netlink.LinkSetMaster(ueTunDevice2, ueVrfDevice); err != nil {
			log.Fatal("[UE][DATA] Unable to set vEth tunnel as slave of VRF interface", err)
			return
		}
	*/

	if err := netlink.LinkSetUp(ueVrfDevice); err != nil {
		log.Fatal("[UE][DATA] Unable to set interface VRF UP", err)
		return
	}
	pduSession.SetVrfDevice(ueVrfDevice)

	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: ueVrfDevice.Attrs().Index,                               // dev gtp-<ECI>
		Scope:     netlink.SCOPE_LINK,                                      // scope n3Inf
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Table:     int(tableId),                                            // table <ECI>
	}

	if err := netlink.RouteReplace(route); err != nil {
		log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}
	pduSession.SetTunRoute(route)

	route = &netlink.Route{
		Dst:       &net.IPNet{IP: netUeIp, Mask: net.CIDRMask(32, 32)}, // default
		LinkIndex: ueVrfDevice.Attrs().Index,                           // dev gtp-<ECI>
		Scope:     netlink.SCOPE_HOST,
		Protocol:  4, // proto static
		Priority:  1, // metric 1
	}

	if err := netlink.RouteReplace(route); err != nil {
		log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond) // Wait for MAC address to be attributed to ueVrfDevice

	links, err := netlink.LinkList()
	var n3Device netlink.Link
	for _, link := range links {
		list, err := netlink.AddrList(link, 0)
		if err != nil {
			log.Fatal(err)
		}
		for _, ip := range list {
			if ip.Contains(gnbIp) {
				n3Device = link
			}
			if ip.Contains(netUeIp) {

				neigh := &netlink.Neigh{
					LinkIndex:    ueVrfDevice.Attrs().Index,
					HardwareAddr: link.Attrs().HardwareAddr,
					IP:           netUeIp,
					State:        netlink.NUD_PERMANENT,
				}

				if err := netlink.NeighAdd(neigh); err != nil {
					log.Fatal("[GNB][GTP] Unable to add vrf to neighbour", err)
				}
			}
		}
	}
	if n3Device == nil {
		log.Fatal("[UE][GTP] Unable to determine the N3 interface from ip ", gnbIp)
	}
	route = &netlink.Route{
		Dst:       &net.IPNet{IP: gnbIp, Mask: net.CIDRMask(32, 32)}, // default
		LinkIndex: n3Device.Attrs().Index,                            // dev gtp-<ECI>
		Protocol:  4,                                                 // proto static
		Priority:  1,                                                 // metric 1
	}

	if err := netlink.RouteReplace(route); err != nil {
		log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}

	//cmdAddFar := []string{nameInf, "1", "--action", "2"}
	//cmdAddFar = []string{nameInf, "2", "--action", "2", "--hdr-creation", "0", fmt.Sprint(gnbPduSession.GetTeidUplink()), msg.UpfIp, "2152"}
	//cmdAddPdr := []string{nameInf, "1", "--pcd", "1", "--hdr-rm", "0", "--ue-ipv4", ueIp, "--f-teid", fmt.Sprint(gnbPduSession.GetTeidDownlink()), msg.GnbIp, "--far-id", "1"}
	//cmdAddPdr = []string{nameInf, "2", "--pcd", "2", "--ue-ipv4", ueIp, "--far-id", "2"}

	qer := ebpf.QerInfo{GateStatusUL: 0, GateStatusDL: 0, Qfi: 5, MaxBitrateUL: 1000000, MaxBitrateDL: 100000, StartUL: 0, StartDL: 0}
	qerId, err := bpfObjects.NewQer(qer)
	if err != nil {
		log.Fatalf("Could not add FAR: %s", err.Error())
	}

	// Downlink
	farDownlink := ebpf.FarInfo{
		Action:                2,
		OuterHeaderCreation:   0,
		Teid:                  gnbPduSession.GetTeidDownlink(),
		RemoteIP:              binary.LittleEndian.Uint32(upfIp),
		LocalIP:               binary.LittleEndian.Uint32(gnbIp),
		IfIndex:               uint32(n3Device.Attrs().Index),
		TransportLevelMarking: 0,
	}
	farDownlinkId, err := bpfObjects.NewFar(farDownlink)
	if err != nil {
		log.Fatalf("Could not add FAR: %s", err.Error())
	}

	pdrDownlink := ebpf.PdrInfo{
		OuterHeaderRemoval: 0,
		FarId:              farDownlinkId,
		QerId:              qerId,
		SdfFilter:          nil,
	}

	bpfObjects.PutPdrUplink(gnbPduSession.GetTeidDownlink(), pdrDownlink)

	// Uplink
	farUplink := ebpf.FarInfo{
		Action:                2,
		OuterHeaderCreation:   0x01,
		Teid:                  gnbPduSession.GetTeidUplink(),
		RemoteIP:              binary.LittleEndian.Uint32(upfIp),
		LocalIP:               binary.LittleEndian.Uint32(gnbIp),
		IfIndex:               uint32(n3Device.Attrs().Index),
		TransportLevelMarking: 0,
	}
	farUplinkId, err := bpfObjects.NewFar(farUplink)
	if err != nil {
		log.Fatalf("Could not add FAR: %s", err.Error())
	}

	pdrUplink := ebpf.IpEntrypointPdrInfo{
		OuterHeaderRemoval: 0,
		FarId:              farUplinkId,
		QerId:              qerId,
	}

	bpfObjects.IpEntrypointObjects.PdrMapDownlinkIp4.Put(netUeIp.To4(), unsafe.Pointer(&pdrUplink))

	for _, ifname := range []string{n3Device.Attrs().Name, ueVrfInf} {
		AttachXdpProgram(ifname, bpfObjects)
	}
	log.Warn(fmt.Sprintf("[UE][GTP] You are using PacketRusher's eBPF Preview mode:"))
	log.Info(fmt.Sprintf("[UE][GTP] Interface %s has successfully been configured for UE %s", ueVrfInf, ueIp))
	log.Info(fmt.Sprintf("[UE][GTP] You can do traffic for this UE using VRF %s, eg:", ueVrfInf))
	log.Info(fmt.Sprintf("[UE][GTP] sudo ip vrf exec %s iperf3 -c IPERF_SERVER -p PORT -t 9000", ueVrfInf))
}

func AttachXdpProgram(ifaceName string, bpfObjects *ebpf.BpfObjects) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("Lookup network iface %q: %s", ifaceName, err.Error())
	}

	go func() {
		// Attach the program.
		_, err = link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     link.XDPGenericMode,
		})
		if err != nil {
			log.Fatalf("Could not attach XDP program: %s", err.Error())
		}
		log.Infof("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
		ch := make(chan string)
		<-ch
	}()

	_ = ebpf.UpfXdpActionStatistic{
		BpfObjects: bpfObjects,
	}

	cmd := exec.Command("ethtool", "-K", ifaceName, "rx", "off", "tx", "off", "sg", "off", "tso", "off", "gso", "off", "gro", "off", "lro", "off")
	if err = cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
