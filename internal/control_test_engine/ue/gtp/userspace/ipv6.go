/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package userspace

import (
	"encoding/binary"
	"net/netip"
)

// routerSolicitation returns the packet, after Headroom free bytes, by which the UE
// at the link-local address source asks the UPF for its prefix (RFC 4861 §4.1).
func routerSolicitation(source netip.Addr) []byte {
	b := make([]byte, Headroom+48)
	packet := b[Headroom:]
	packet[0], packet[5], packet[6], packet[7] = 0x60, 8, 58, 255 // 8 bytes of ICMPv6, hop limit 255
	copy(packet[8:], source.AsSlice())
	copy(packet[24:], netip.MustParseAddr("ff02::2").AsSlice()) // all routers
	packet[40] = 133
	binary.BigEndian.PutUint16(packet[42:], checksum(packet))
	return b
}

// advertisedPrefix reports whether packet is a Router Advertisement and, if a valid
// one, the /64 prefix it gives the UE to configure its address (RFC 4861 §4.2,
// §4.6.2). Only ICMPv6 that directly follows the IPv6 header is looked at: a
// fragment or a packet with extension headers is ordinary traffic.
func advertisedPrefix(packet []byte) (prefix netip.Prefix, advertisement bool) {
	if len(packet) <= 40 || packet[0]>>4 != 6 || packet[6] != 58 || packet[40] != 134 {
		return prefix, false
	}
	if len(packet) < 56 || len(packet)%8 != 0 || packet[7] != 255 || checksum(packet) != 0 {
		return prefix, true
	}
	options := packet[56:]
	for len(options) != 0 && options[1] != 0 && 8*int(options[1]) <= len(options) {
		// Prefix Information: 32 bytes, a /64 for autoconfiguration, still valid.
		if options[0] == 3 && options[1] == 4 && options[2] == 64 && options[3]&0x40 != 0 && binary.BigEndian.Uint32(options[4:]) != 0 {
			return netip.PrefixFrom(netip.AddrFrom16([16]byte(options[16:32])), 64), true
		}
		options = options[8*int(options[1]):]
	}
	return prefix, true
}

// checksum computes the ICMPv6 checksum of an IPv6 packet of even length without
// extension headers (RFC 4443 §2.3); it is zero if the packet carries a correct one.
func checksum(packet []byte) uint16 {
	sum := uint32(len(packet)-40) + 58 // pseudo-header: ICMPv6 length and next header
	for i := 8; i < len(packet); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(packet[i:])) // addresses, then ICMPv6
	}
	for sum > 0xffff {
		sum = sum>>16 + sum&0xffff
	}
	return ^uint16(sum)
}
