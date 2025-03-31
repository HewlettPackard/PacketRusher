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

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
)

func CaptureTraffic(path cli.Path) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	n2Link := findN2Link()

	pcapw := pcapgo.NewWriter(f)
	if err := pcapw.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
		log.Fatalf("WriteFileHeader: %v", err)
	}

	handle, err := pcapgo.NewEthernetHandle(n2Link.Attrs().Name)
	if err != nil {
		log.Fatalf("NewEthernetHandle: %v", err)
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

func findN2Link() netlink.Link {
	config := config.GetConfig()
	ip := net.ParseIP(config.GNodeB.ControlIF.Ip)

	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Unable to capture traffic to AMF as we are unable to get Links information %s", err)
	}

	for _, link := range links {
		addrs, err := netlink.AddrList(link, 0)
		if err != nil {
			log.Errorf("Unable to get IPs of link %s: %s", link.Attrs().Name, err)
			continue
		}

		for _, addr := range addrs {
			if addr.IP.Equal(ip) {
				return link
			}
		}
	}

	log.Fatalf("Unable to find network interface providing gNodeB IP %s", ip)
	return nil // unreachable
}
