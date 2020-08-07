package pfcpType

import (
	"fmt"
	"math/bits"
)

const (
	PDNTypeIpv4 uint8 = iota + 1
	PDNTypeIpv6
	PDNTypeIpv4v6
	PDNTypeNonIp
	PDNTypeEthernet
)

type PDNType struct {
	PdnType uint8 // 0x00000111
}

func (p *PDNType) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(p.PdnType) > 3 {
		return []byte(""), fmt.Errorf("PDN type shall not be greater than 3 bits binary integer")
	}
	data = append([]byte(""), byte(p.PdnType))

	return data, nil
}

func (p *PDNType) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	p.PdnType = uint8(data[idx]) & Mask3
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
