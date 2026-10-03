// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	gPDU                = 255
	pduSessionContainer = 0x85
)

// Encode includes an uplink PDU Session Container when a QoS flow is supplied.
func Encode(teid uint32, qfi uint8, payload []byte) ([]byte, error) {
	if teid == 0 || qfi > 63 {
		return nil, errors.New("TEID must be nonzero and QFI must be 0..63")
	}
	header := 8
	if qfi != 0 {
		header = 16
	}
	if len(payload) > 65507-header {
		return nil, errors.New("GTP-U payload exceeds UDP datagram limit")
	}
	b := make([]byte, header+len(payload))
	b[0], b[1] = 0x30, gPDU
	binary.BigEndian.PutUint16(b[2:4], uint16(len(b)-8))
	binary.BigEndian.PutUint32(b[4:8], teid)
	if qfi != 0 {
		b[0] = 0x34
		b[11], b[12], b[13], b[14] = pduSessionContainer, 1, 0x10, qfi
	}
	copy(b[header:], payload)
	return b, nil
}

// Decode validates optional fields and every extension before exposing a T-PDU.
func Decode(b []byte) (uint32, []byte, error) {
	if len(b) < 8 || b[0]>>5 != 1 || b[0]&0x10 == 0 || b[0]&8 != 0 || b[1] != gPDU {
		return 0, nil, errors.New("not a GTPv1-U T-PDU")
	}
	if int(binary.BigEndian.Uint16(b[2:4])) != len(b)-8 {
		return 0, nil, errors.New("GTP-U length differs from datagram")
	}
	teid := binary.BigEndian.Uint32(b[4:8])
	if teid == 0 {
		return 0, nil, errors.New("zero TEID")
	}
	offset := 8
	if b[0]&7 != 0 {
		if len(b) < 12 {
			return 0, nil, errors.New("truncated GTP-U optional fields")
		}
		offset = 12
		next := b[11]
		if b[0]&4 == 0 && next != 0 {
			return 0, nil, errors.New("extension flag is absent")
		}
		for extensions := 0; next != 0; extensions++ {
			if extensions >= 32 || offset >= len(b) {
				return 0, nil, errors.New("invalid GTP-U extension chain")
			}
			size := int(b[offset]) * 4
			if size < 4 || size > len(b)-offset {
				return 0, nil, errors.New("invalid GTP-U extension size")
			}
			next = b[offset+size-1]
			offset += size
		}
	}
	if offset == len(b) {
		return 0, nil, errors.New("empty T-PDU")
	}
	if _, _, err := IPAddresses(b[offset:]); err != nil {
		return 0, nil, err
	}
	return teid, b[offset:], nil
}

// IPAddresses rejects truncated packets and inconsistent IP length fields.
func IPAddresses(b []byte) (source, destination []byte, err error) {
	if len(b) == 0 {
		return nil, nil, errors.New("empty IP packet")
	}
	switch b[0] >> 4 {
	case 4:
		if len(b) < 20 || int(b[0]&15)*4 < 20 || int(b[0]&15)*4 > len(b) || int(binary.BigEndian.Uint16(b[2:4])) != len(b) {
			return nil, nil, errors.New("invalid IPv4 packet")
		}
		return b[12:16], b[16:20], nil
	case 6:
		if len(b) < 40 || int(binary.BigEndian.Uint16(b[4:6]))+40 != len(b) {
			return nil, nil, errors.New("invalid IPv6 packet")
		}
		return b[8:24], b[24:40], nil
	default:
		return nil, nil, fmt.Errorf("unsupported IP version %d", b[0]>>4)
	}
}

func echoResponse(b []byte) []byte {
	// Echo requests carry a sequence number and no TEID. Recovery IE reports a
	// stable restart counter for this socket's lifetime; peers can maintain liveness.
	if len(b) < 12 || b[0]>>5 != 1 || b[0]&0x10 == 0 || b[1] != 1 || b[0]&2 == 0 || binary.BigEndian.Uint32(b[4:8]) != 0 || int(binary.BigEndian.Uint16(b[2:4])) != len(b)-8 {
		return nil
	}
	r := []byte{0x32, 2, 0, 6, 0, 0, 0, 0, b[8], b[9], 0, 0, 14, 0}
	return r
}
