/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package pcap

import (
	"my5G-RANTester/config"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func CaptureTraffic(path cli.Path) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	config := config.GetConfig()
	ip := net.ParseIP(config.GNodeB.ControlIF.Ip)

	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Unable to capture traffic to AMF as we are unable to get Links informations %s", err)
	}
	var n2Link *netlink.Link
outer:
	for _, link := range links {
		addrs, err := netlink.AddrList(link, 0)
		if err != nil {
			log.Errorf("Unable to get IPs of link %s: %s", link.Attrs().Name, err)
			continue
		}
		for _, addr := range addrs {
			if addr.IP.Equal(ip) {
				n2Link = &link
				break outer
			}
		}
	}

	if n2Link == nil {
		log.Fatalf("Unable to find network interface providing gNodeB IP %s", ip)
	}

	pcapw := pcapgo.NewWriter(f)
	if err := pcapw.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
		log.Fatalf("WriteFileHeader: %v", err)
	}

	handle, err := pcapgo.NewEthernetHandle((*n2Link).Attrs().Name)
	if err != nil {
		log.Fatalf("OpenEthernet: %v", err)
	}

	pkgsrc := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	go func() {
		for packet := range pkgsrc.Packets() {
			if err := pcapw.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
				log.Fatalf("pcap.WritePacket(): %v", err)
			}
		}
	}()
}
