package pfcpType

import (
	"fmt"
)

type PFCPSMReqFlags struct {
	Qaurr bool
	Sndem bool
	Drobu bool
}

func (p *PFCPSMReqFlags) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(p.Qaurr)<<2 |
		btou(p.Sndem)<<1 |
		btou(p.Drobu)
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (p *PFCPSMReqFlags) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	p.Qaurr = utob(uint8(data[idx]) & BitMask3)
	p.Sndem = utob(uint8(data[idx]) & BitMask2)
	p.Drobu = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
