package service

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
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

	// Set IP interface up
	if err := netlink.LinkSetUp(newInterface); err != nil {
		log.Fatal("[UE][DATA] Error in setting virtual interface up ", err)
	}

	if err := netlink.AddrAdd(newInterface, addrTun); err != nil {
		log.Fatal("[UE][DATA] Error in adding IP for virtual interface", err)
	}

	fmt.Println("[UE][DATA] UE is ready for using data plane")

	defer func() {
		_ = netlink.LinkSetDown(newInterface)
		_ = netlink.LinkDel(newInterface)
	}()

	time.Sleep(60 * time.Minute)

}
