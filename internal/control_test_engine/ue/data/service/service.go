package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"time"
)

func InitDataPlane(ue *context.UEContext, message []byte) {

	// get UE GNB IP.
	ue.SetGnbIp(message)

	// create interface for data plane.
	gatewayIp := ue.GetGatewayIp()
	ueIp := ue.GetIp()
	ueGnbIp := ue.GetGnbIp()
	nameInf := fmt.Sprintf("uetun%d", ue.GetPduSesssionId())

	newInterface := &netlink.Iptun{
		LinkAttrs: netlink.LinkAttrs{
			Name: nameInf,
		},
		Local:  ueGnbIp,
		Remote: gatewayIp,
	}

	if err := netlink.LinkAdd(newInterface); err != nil {
		log.Fatal(err)
	}

	// add an IP address to a link device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ueIp).To4(),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
	}

	if err := netlink.AddrAdd(newInterface, addrTun); err != nil {
		log.Fatal("[UE][DATA] Error in adding IP for virtual interface", err)
	}

	// Set IP interface up
	if err := netlink.LinkSetUp(newInterface); err != nil {
		log.Fatal("[UE][DATA] Error in setting virtual interface up ", err)
	}

	ueRoute := &netlink.Route{
		LinkIndex: newInterface.Attrs().Index,
		Dst: &net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Gw:    net.ParseIP(ueIp).To4(),
		Table: 1,
	}

	if err := netlink.RouteAdd(ueRoute); err != nil {
		log.Fatal("[UE][DATA] Error in setting route", err)
	}

	log.Info("[UE][DATA] UE is ready for using data plane")

	defer func() {
		_ = netlink.LinkSetDown(newInterface)
		_ = netlink.LinkDel(newInterface)
		_ = netlink.RouteDel(ueRoute)
	}()

	time.Sleep(60 * time.Minute)

}
