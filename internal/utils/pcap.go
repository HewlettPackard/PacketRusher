/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package pcap

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"net"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func CaptureTraffic(path cli.Path) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	config, err := config.GetConfig()
	ip := net.ParseIP(config.AMF.Ip)
	route, err := netlink.RouteGet(ip)
	if err != nil || len(route) <= 0 {
		log.Fatal("Unable to capture traffic to AMF as we are unable to determine route to AMF. Aborting.", err)
	}

	iif, err := netlink.LinkByIndex(route[0].LinkIndex)

	pcapw := pcapgo.NewWriter(f)
	if err := pcapw.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
		log.Fatalf("WriteFileHeader: %v", err)
	}

	handle, err := pcapgo.NewEthernetHandle(iif.Attrs().Name)
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
