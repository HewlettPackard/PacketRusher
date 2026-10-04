/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package gtp

import (
	"fmt"
	"net/netip"

	"github.com/vishvananda/netlink"
)

// overhead is what GTP-U adds to a packet on N3: IPv4 (20), UDP (8), the GTP-U header
// (8), its optional fields (4) and the PDU Session Container carrying the QFI (4).
const overhead = 44

// tunnelMTU returns the MTU of the tunnels of an N3 interface whose MTU is n3: what
// the overhead leaves, or ue.tunnelmtu if configured, to match the UPF or a path
// that carries less.
func tunnelMTU(n3, configured int, ipv6 bool) (int, error) {
	highest := min(n3, 65535) - overhead // Loopback's MTU exceeds the largest IPv4 packet.
	lowest := 68
	if ipv6 {
		lowest = 1280
	}
	mtu := highest
	if configured != 0 {
		mtu = configured
	}
	if mtu < lowest || mtu > highest {
		return 0, fmt.Errorf("tunnel MTU %d is not between %d and the %d that the N3 MTU %d leaves", mtu, lowest, highest, n3)
	}
	return mtu, nil
}

// SetTunnelMTU sets that MTU on the devices of a tunnel, from the interface holding
// its N3 address source: gtp5g's own default only leaves room for 36 bytes.
func SetTunnelMTU(source netip.Addr, configured int, ipv6 bool, devices ...netlink.Link) error {
	addresses, err := netlink.AddrList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	for _, address := range addresses {
		if !address.IP.Equal(source.AsSlice()) {
			continue
		}
		n3, err := netlink.LinkByIndex(address.LinkIndex)
		if err != nil {
			return err
		}
		mtu, err := tunnelMTU(n3.Attrs().MTU, configured, ipv6)
		for _, device := range devices {
			if err == nil {
				err = netlink.LinkSetMTU(device, mtu)
			}
		}
		return err
	}
	return fmt.Errorf("no interface has the N3 address %s", source)
}
