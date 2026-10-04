/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package userspace

import (
	"encoding/binary"
	"errors"
)

// Headroom is the space Encode needs in front of a packet for the GTP-U header.
const Headroom = 16

// Encode writes the G-PDU header of the packet b[Headroom:] in front of it, in
// place, and returns the datagram. A non-zero QFI travels in a PDU Session
// Container extension header (TS 38.415), which 5G UPFs expect on N3.
func Encode(b []byte, teid uint32, qfi uint8) []byte {
	if qfi == 0 {
		b = b[Headroom-8:]
		b[0] = 0x30 // version 1, protocol type GTP
	} else {
		// The same with the E flag, no sequence or N-PDU number, then the container:
		// type 0x85, 4 bytes long, UL PDU SESSION INFORMATION, QFI, end of the chain.
		copy(b, []byte{0x34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x85, 1, 0x10, qfi, 0})
	}
	b[1] = 0xff // G-PDU
	binary.BigEndian.PutUint16(b[2:], uint16(len(b)-8))
	binary.BigEndian.PutUint32(b[4:], teid)
	return b
}

// Decode returns the TEID and the packet of a G-PDU, skipping its extension
// headers. As TS 29.281 §5.1 asks of a receiver, the spare bit is not evaluated,
// nor is the next extension header type unless the E flag is set.
func Decode(b []byte) (teid uint32, packet []byte, err error) {
	if len(b) < 8 || b[0]&0xf0 != 0x30 || b[1] != 0xff {
		return 0, nil, errors.New("not a GTPv1-U G-PDU")
	}
	end, offset := 8+int(binary.BigEndian.Uint16(b[2:])), 8
	if b[0]&0x07 != 0 { // E, S or PN: the three optional fields are all present
		offset = 12
	}
	if offset > end || end > len(b) {
		return 0, nil, errors.New("truncated G-PDU")
	}
	if b[0]&0x04 != 0 {
		for next := b[11]; next != 0; next = b[offset-1] {
			if offset >= end || b[offset] == 0 || offset+4*int(b[offset]) > end {
				return 0, nil, errors.New("malformed GTP-U extension header")
			}
			offset += 4 * int(b[offset])
		}
	}
	return binary.BigEndian.Uint32(b[4:]), b[offset:end], nil
}

// echoResponse answers an Echo Request, as the UPF may supervise the N3 path with
// them (TS 29.281 §7.2); it returns nil for any other message.
func echoResponse(b []byte) []byte {
	if len(b) < 12 || b[0]&0xf2 != 0x32 || b[1] != 1 {
		return nil
	}
	// Same sequence number, and the mandatory Recovery IE with its fixed value of 0.
	return []byte{0x32, 2, 0, 6, 0, 0, 0, 0, b[8], b[9], 0, 0, 14, 0}
}
