// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"strconv"
)

// GNBIdentity is the numeric gNB identifier and its cell suffix in a 36-bit NCI.
// Zero BitLength in configuration preserves the original 24-bit identifier.
type GNBIdentity struct {
	Value     uint64
	BitLength uint8
	Cell      uint16
}

func ParseGNBIdentity(id string, bits uint8, cell uint16) (GNBIdentity, error) {
	if bits == 0 {
		bits = 24
	}
	if bits < 22 || bits > 32 {
		return GNBIdentity{}, fmt.Errorf("gnbidlength must be between 22 and 32 bits")
	}
	if id == "" {
		return GNBIdentity{}, fmt.Errorf("gnbid must contain hexadecimal digits")
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return GNBIdentity{}, fmt.Errorf("gnbid %q must contain hexadecimal digits without a prefix", id)
		}
	}
	value, err := strconv.ParseUint(id, 16, 32)
	if err != nil || value >= uint64(1)<<bits {
		return GNBIdentity{}, fmt.Errorf("gnbid %q does not fit in %d bits", id, bits)
	}
	if uint64(cell) >= uint64(1)<<(36-bits) {
		return GNBIdentity{}, fmt.Errorf("cellid %d does not fit in the %d-bit cell suffix", cell, 36-bits)
	}
	return GNBIdentity{Value: value, BitLength: bits, Cell: cell}, nil
}

func (id GNBIdentity) String() string {
	return fmt.Sprintf("%0*X", (int(id.BitLength)+3)/4, id.Value)
}

// Bytes encodes a left-aligned ASN.1 BIT STRING, including zero padding bits.
func (id GNBIdentity) Bytes() []byte {
	return identityBytes(id.Value, id.BitLength)
}

func (id GNBIdentity) CellBytes() []byte {
	return identityBytes(id.Value<<uint(36-id.BitLength)|uint64(id.Cell), 36)
}

func identityBytes(value uint64, bits uint8) []byte {
	out := make([]byte, (int(bits)+7)/8)
	value <<= uint(len(out)*8 - int(bits))
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = byte(value)
		value >>= 8
	}
	return out
}

func (p PlmnList) Identity() (GNBIdentity, error) {
	return ParseGNBIdentity(p.GnbId, p.GnbIDLength, p.CellID)
}

// GNBIDAt checks the full offset before any gNB network connection is started.
func (p PlmnList) GNBIDAt(offset int) (string, error) {
	id, err := p.Identity()
	if err != nil {
		return "", err
	}
	if offset < 0 || uint64(offset) >= (uint64(1)<<id.BitLength)-id.Value {
		return "", fmt.Errorf("gNB offset %d exceeds the configured %d-bit identifier range", offset, id.BitLength)
	}
	id.Value += uint64(offset)
	return id.String(), nil
}
