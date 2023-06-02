package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	gtpLink "my5G-RANTester/internal/cmd/gogtp5g-link"
	"net"
	"time"
)

func SetupGtpInterface(ue *context.UEContext, message []byte) {
    // get UE GNB IP.
    ue.SetGnbIp(net.ParseIP(string(message)))

	// create interface for data plane.
	/*gatewayIp := ue.GetGatewayIp()
	*/

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

    log.Info(fmt.Sprintf("[GNB][GTP] Interface %s has successfully been configured for UE %s", nameInf, ueIp))

/*
    // add a tunnel by giving GTP peer's IP, subscriber's IP,
   if err := gnb.GetN3Plane().AddTunnelOverride(
    	net.ParseIP("30.103.12.22"), // GTP peer's IP
    	net.ParseIP("30.103.12.26"),     // subscriber's IP
    	ue.GetTeidUplink(),                 // outgoing TEID
    	ue.GetTeidDownlink(),                 // incoming TEID
    ); err != nil {
        log.Info("[GNB][NGAP][UE] Err: ", err)
    }

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
    }

	// add an IP address to a link device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP("30.103.12.26").To4(),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
	}

	if err := netlink.AddrAdd(gnb.GetN3Plane().KernelGTP.Link, addrTun); err != nil {
		log.Info("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}*/

}