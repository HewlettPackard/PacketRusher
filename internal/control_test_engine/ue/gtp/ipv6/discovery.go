// SPDX-License-Identifier: Apache-2.0
// Package ipv6 implements the user-plane IPv6 prefix discovery required after
// NAS allocates an interface identifier (TS 23.501 5.8.2.2.3).
package ipv6

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

type Advertisement struct {
	Address           netip.Addr
	Prefix            netip.Prefix
	Router            netip.Addr
	ValidLifetime     uint32
	PreferredLifetime uint32
	RouterLifetime    uint16
}

func LinkLocal(iid [8]byte) netip.Addr {
	var bytes [16]byte
	bytes[0], bytes[1] = 0xfe, 0x80
	copy(bytes[8:], iid[:])
	return netip.AddrFrom16(bytes)
}

// RouterSolicitation returns a complete IPv6 packet. The GTP-U transport sends
// this directly, independent of host kernel RA/SLAAC sysctl settings.
func RouterSolicitation(iid [8]byte) []byte {
	b := make([]byte, 48)
	b[0], b[6], b[7] = 0x60, 58, 255
	binary.BigEndian.PutUint16(b[4:6], 8)
	source, destination := LinkLocal(iid).As16(), netip.MustParseAddr("ff02::2").As16()
	copy(b[8:24], source[:])
	copy(b[24:40], destination[:])
	b[40] = 133
	binary.BigEndian.PutUint16(b[42:44], checksum(b))
	return b
}

// ParseAdvertisement validates RFC 4861 6.1.2 and the /64 autonomous prefix
// rules from RFC 4862 5.5.3. Fragmented or extended-header advertisements are
// deliberately rejected; RFC 6980 prohibits fragmented Neighbor Discovery.
func ParseAdvertisement(packet []byte, iid [8]byte) (Advertisement, error) {
	var result Advertisement
	if len(packet) < 56 || packet[0]>>4 != 6 || packet[6] != 58 || packet[7] != 255 || int(binary.BigEndian.Uint16(packet[4:6])) != len(packet)-40 || packet[40] != 134 || packet[41] != 0 {
		return result, fmt.Errorf("invalid IPv6 Router Advertisement envelope")
	}
	source := netip.AddrFrom16([16]byte(packet[8:24]))
	destination := netip.AddrFrom16([16]byte(packet[24:40]))
	if !source.IsLinkLocalUnicast() || (destination != netip.MustParseAddr("ff02::1") && destination != LinkLocal(iid)) || checksum(packet) != 0 {
		return result, fmt.Errorf("invalid IPv6 Router Advertisement source, destination, or checksum")
	}
	for offset := 56; offset < len(packet); {
		if len(packet)-offset < 2 || packet[offset+1] == 0 {
			return result, fmt.Errorf("invalid Router Advertisement option length")
		}
		length := int(packet[offset+1]) * 8
		if length > len(packet)-offset {
			return result, fmt.Errorf("truncated Router Advertisement option")
		}
		option := packet[offset : offset+length]
		if option[0] == 3 {
			if length != 32 {
				return result, fmt.Errorf("invalid Prefix Information option length")
			}
			valid, preferred := binary.BigEndian.Uint32(option[4:8]), binary.BigEndian.Uint32(option[8:12])
			prefix := netip.PrefixFrom(netip.AddrFrom16([16]byte(option[16:32])), int(option[2])).Masked()
			if option[2] == 64 && option[3]&0x40 != 0 && preferred <= valid && !prefix.Addr().IsLinkLocalUnicast() && !prefix.Addr().IsMulticast() && !prefix.Addr().IsUnspecified() && !result.Address.IsValid() {
				address := prefix.Addr().As16()
				copy(address[8:], iid[:])
				result = Advertisement{Address: netip.AddrFrom16(address), Prefix: prefix, Router: source, ValidLifetime: valid, PreferredLifetime: preferred, RouterLifetime: binary.BigEndian.Uint16(packet[46:48])}
			}
		}
		offset += length
	}
	if !result.Address.IsValid() {
		return result, fmt.Errorf("Router Advertisement has no usable autonomous /64 prefix")
	}
	return result, nil
}

func checksum(packet []byte) uint16 {
	var sum uint32
	for offset := 8; offset < 40; offset += 2 {
		sum += uint32(binary.BigEndian.Uint16(packet[offset : offset+2]))
	}
	length := uint32(len(packet) - 40)
	sum += length>>16 + length&0xffff + 58
	for offset := 40; offset+1 < len(packet); offset += 2 {
		sum += uint32(binary.BigEndian.Uint16(packet[offset : offset+2]))
	}
	if len(packet)&1 != 0 {
		sum += uint32(packet[len(packet)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = sum&0xffff + sum>>16
	}
	return ^uint16(sum)
}
