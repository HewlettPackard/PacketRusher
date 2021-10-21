package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
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
		log.Info("UE][DATA] Error in setting virtual interface", err)
		return
	}

	// add an IP address to a link device.
	addrTun := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP(ueIp).To4(),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
	}

	if err := netlink.AddrAdd(newInterface, addrTun); err != nil {
		log.Info("[UE][DATA] Error in adding IP for virtual interface", err)
		return
	}

	// Set IP interface up
	if err := netlink.LinkSetUp(newInterface); err != nil {
		log.Info("[UE][DATA] Error in setting virtual interface up ", err)
		return
	}

	// create route in linux to table 1
	ueRoute := &netlink.Route{
		LinkIndex: newInterface.Attrs().Index,
		Src:       net.ParseIP(ueIp).To4(),
		Dst: &net.IPNet{
			IP:   net.IPv4zero,
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Table: int(ue.GetPduSesssionId()),
	}

	if err := netlink.RouteAdd(ueRoute); err != nil {
		log.Info("[UE][DATA] Error in setting route", err)
		return
	}

	// create rule to mapped traffic
	ueRule := netlink.NewRule()

	ueRule.Src = &net.IPNet{
		IP:   net.ParseIP(ueIp).To4(),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}

	ueRule.Table = int(ue.GetPduSesssionId())

	if err := netlink.RuleAdd(ueRule); err != nil {
		log.Info("[UE][DATA] Error in setting rule", err)
		return
	}

	log.Info("[UE][DATA] UE is ready for using data plane")

	// contex of tun interface
	ue.SetTunInterface(newInterface)
	ue.SetTunRoute(ueRoute)
	ue.SetTunRule(ueRule)

}
