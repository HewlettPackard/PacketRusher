package pfcpType

import (
	"fmt"
	"math/bits"
)

type GateStatus struct {
	UlGate uint8 // 0x00001100
	DlGate uint8 // 0x00000011
}

func (g *GateStatus) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(g.UlGate) > 2 {
		return []byte(""), fmt.Errorf("UL gate shall not be greater than 2 bits binary integer")
	}
	if bits.Len8(g.DlGate) > 2 {
		return []byte(""), fmt.Errorf("DL gate shall not be greater than 2 bits binary integer")
	}
	tmpUint8 := g.UlGate<<2 | g.DlGate
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (g *GateStatus) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	g.UlGate = uint8(data[idx]) >> 2 & Mask2
	g.DlGate = uint8(data[idx]) & Mask2
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
