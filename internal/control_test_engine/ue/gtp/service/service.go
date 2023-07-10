package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
    gnbContext 	"my5G-RANTester/internal/control_test_engine/gnb/context"
	gtpLink "my5G-RANTester/internal/cmd/gogtp5g-link"
	gtpTunnel "my5G-RANTester/internal/cmd/gogtp5g-tunnel"
	"github.com/vishvananda/netlink"

    "encoding/json"
	"net"
	"time"
	"strings"
)

func SetupGtpInterface(ue *context.UEContext, message []byte) {
    var msg gnbContext.UEMessage
    jsonErr := json.Unmarshal(message, &msg)
    if jsonErr != nil {
        log.Fatal(jsonErr)
    }

    // get UE GNB IP.
    ue.SetGnbIp(net.ParseIP(msg.GetGnbIp()))

	ueGnbIp := ue.GetGnbIp()
    ueIp := ue.GetIp()
    msin := ue.GetMsin()
	nameInf := fmt.Sprintf("val%s", msin)

   _ = gtpLink.CmdDel(nameInf)

    go func() {
        if err := gtpLink.CmdAdd(nameInf, 1, ueGnbIp.String()); err != nil {
            log.Fatal("[GNB][GTP] Unable to create Kernel GTP interface: ", err, msin, nameInf)
            return
        }
    }()

    time.Sleep(time.Second)

	cmdAddFar := []string{nameInf, "1", "--action", "2"}
	log.Info("[UE][GTP] ", strings.Join(cmdAddFar, " "))
    if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
        log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
        return
    }

    cmdAddFar = []string{nameInf, "2", "--action", "2", "--hdr-creation", "0", msg.GetOTeid(), msg.GetUpfIp(), "2152"}
	log.Info("[UE][GTP] ", strings.Join(cmdAddFar, " "))
    if err := gtpTunnel.CmdAddFAR(cmdAddFar); err != nil {
        log.Fatal("[UE][GTP] Unable to create FAR ", err)
        return
    }

    cmdAddPdr := []string{nameInf, "1", "--pcd", "1", "--hdr-rm", "1", "--ue-ipv4", ueIp, "--f-teid", msg.GetITeid(), msg.GetGnbIp(), "--far-id", "1"}
    log.Info("[UE][GTP] ", strings.Join(cmdAddPdr, " "))

    if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
        log.Fatal("[GNB][GTP] Unable to create FAR: ", err)
        return
    }

    cmdAddPdr = []string{nameInf, "2", "--pcd", "2", "--ue-ipv4", ueIp, "--far-id", "2"}
	log.Info("[UE][GTP] ", strings.Join(cmdAddPdr, " "))
    if err := gtpTunnel.CmdAddPDR(cmdAddPdr); err != nil {
        log.Fatal("[UE][GTP] Unable to create FAR ", err)
        return
    }

	// add an IP address to a link device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ueIp).To4(),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
	}

	link, _ := netlink.LinkByName(nameInf)
	ue.SetTunInterface(link)

	if err := netlink.AddrAdd(link, addrTun); err != nil {
		log.Info("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}

    log.Info(fmt.Sprintf("[GNB][GTP] Interface %s has successfully been configured for UE %s", nameInf, ueIp))

/*
	route := &netlink.Route{
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, // default
		LinkIndex: gnb.GetN3Plane().KernelGTP.Link.Attrs().Index,            // dev gtp-<ECI>
		Scope:     netlink.SCOPE_LINK,                                      // scope link
		Protocol:  4,                                                       // proto static
		Priority:  1,                                                       // metric 1
		Table:     1001,                                                    // table <ECI>
	}

	if netlink.RouteReplace(route) != nil {
	   log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
	}

    rule := netlink.NewRule()
    rule.IifName = gnb.GetN3Plane().KernelGTP.Link.Name
    rule.Table = 1001

    if err := netlink.RuleAdd(rule); err != nil {
	   log.Fatal("[GNB][GTP] Unable to create Kernel Route ", err)
    }*/

}
