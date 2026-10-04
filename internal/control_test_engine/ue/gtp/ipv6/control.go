// SPDX-License-Identifier: Apache-2.0
package ipv6

// routerAdvertisement also recognizes ND hidden behind IPv6 extension headers.
// Prefix discovery rejects such envelopes; delivering them to TUN instead could
// let host SLAAC bypass the per-session routing and allocated-interface policy.
func RouterAdvertisement(packet []byte) bool {
	if len(packet) < 40 || packet[0]>>4 != 6 {
		return false
	}
	next, offset := packet[6], 40
	for count := 0; count < 8; count++ {
		switch next {
		case 58:
			return offset < len(packet) && packet[offset] == 134
		case 0, 43, 60, 51:
			if len(packet)-offset < 2 {
				return true
			}
			length := (int(packet[offset+1]) + 1) * 8
			if next == 51 {
				length = (int(packet[offset+1]) + 2) * 4
			}
			if length > len(packet)-offset {
				return true
			}
			next = packet[offset]
			offset += length
		case 44:
			if len(packet)-offset < 8 {
				return true
			}
			// RFC 6980 forbids fragmented ND. Later ICMPv6 fragments do not
			// identify the message type, so none reach host ND processing.
			if packet[offset] == 58 {
				return true
			}
			next = packet[offset]
			offset += 8
		default:
			return false
		}
	}
	return true // Excessive or unterminated extension chain is not host input.
}
