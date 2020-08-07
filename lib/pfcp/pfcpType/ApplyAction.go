package pfcpType

import (
	"fmt"
)

type ApplyAction struct {
	Dupl bool
	Nocp bool
	Buff bool
	Forw bool
	Drop bool
}

func (a *ApplyAction) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(a.Dupl)<<4 |
		btou(a.Nocp)<<3 |
		btou(a.Buff)<<2 |
		btou(a.Forw)<<1 |
		btou(a.Drop)
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (a *ApplyAction) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	a.Dupl = utob(uint8(data[idx]) & BitMask5)
	a.Nocp = utob(uint8(data[idx]) & BitMask4)
	a.Buff = utob(uint8(data[idx]) & BitMask3)
	a.Forw = utob(uint8(data[idx]) & BitMask2)
	a.Drop = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
