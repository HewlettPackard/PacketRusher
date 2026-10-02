// SPDX-License-Identifier: Apache-2.0
package gtp

import (
	"fmt"
	"net/netip"

	"github.com/vishvananda/netlink"
)

// Uplink packets include IPv4 (20), UDP (8), the base GTP-U header (8),
// optional GTP fields (4), and the PDU Session Container carrying QFI (4).
// Reserve the full overhead even when a session has no QFI.
const IPv4GTPOverhead = 44

func payloadMTU(underlay, configured int) (int, error) {
	if underlay > 65535 {
		underlay = 65535 // IPv4's total-length field is 16 bits.
	}
	maximum := underlay - IPv4GTPOverhead
	if maximum < 68 {
		return 0, fmt.Errorf("N3 MTU %d cannot carry an IPv4 tunnel with %d bytes of overhead", underlay, IPv4GTPOverhead)
	}
	if configured == 0 {
		return maximum, nil
	}
	if configured < 68 || configured > maximum {
		return 0, fmt.Errorf("ue.tunnelmtu %d must be between 68 and %d for N3 MTU %d", configured, maximum, underlay)
	}
	return configured, nil
}

type mtuOperations struct {
	links     func() ([]netlink.Link, error)
	addresses func(netlink.Link, int) ([]netlink.Addr, error)
	set       func(netlink.Link, int) error
}

// SetTunnelMTU overrides gtp5g's default, which subtracts only 36 bytes.
// Use the interface holding the configured N3 source address, including aliases.
// A smaller explicit value accommodates a lower path MTU or matches the UPF.
func SetTunnelMTU(tunnel netlink.Link, source netip.Addr, configured int) error {
	return setTunnelMTU(tunnel, source, configured, mtuOperations{
		links: netlink.LinkList, addresses: netlink.AddrList, set: netlink.LinkSetMTU,
	})
}

func setTunnelMTU(tunnel netlink.Link, source netip.Addr, configured int, ops mtuOperations) error {
	if tunnel == nil || tunnel.Attrs() == nil {
		return fmt.Errorf("tunnel interface is missing")
	}
	source = source.Unmap()
	if !source.Is4() {
		return fmt.Errorf("N3 source %s must be IPv4", source)
	}
	links, err := ops.links()
	if err != nil {
		return fmt.Errorf("list N3 interfaces: %w", err)
	}
	underlay := 0
	for _, link := range links {
		if link.Attrs() == nil || link.Attrs().Index == tunnel.Attrs().Index {
			continue
		}
		addresses, err := ops.addresses(link, netlink.FAMILY_V4)
		if err != nil {
			return fmt.Errorf("list addresses on %s: %w", link.Attrs().Name, err)
		}
		for _, address := range addresses {
			if address.IPNet == nil {
				continue
			}
			ip, ok := netip.AddrFromSlice(address.IP)
			if ok && ip.Unmap() == source && (underlay == 0 || link.Attrs().MTU < underlay) {
				underlay = link.Attrs().MTU
			}
		}
	}
	if underlay == 0 {
		return fmt.Errorf("N3 source %s is not assigned to an interface with a usable MTU", source)
	}
	mtu, err := payloadMTU(underlay, configured)
	if err != nil {
		return err
	}
	if err := ops.set(tunnel, mtu); err != nil {
		return fmt.Errorf("set MTU %d on %s: %w", mtu, tunnel.Attrs().Name, err)
	}
	tunnel.Attrs().MTU = mtu
	return nil
}
