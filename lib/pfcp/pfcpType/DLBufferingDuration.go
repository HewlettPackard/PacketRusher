package pfcpType

import (
	"fmt"
	"math/bits"
)

type DLBufferingDuration struct {
	TimerUnit  uint8 // 0x11100000
	TimerValue uint8 // 0x00011111
}

func (d *DLBufferingDuration) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(d.TimerUnit) > 3 {
		return []byte(""), fmt.Errorf("Timer unit shall not be greater than 3 bits binary integer")
	}
	if bits.Len8(d.TimerValue) > 5 {
		return []byte(""), fmt.Errorf("Timer data shall not be greater than 5 bits binary integer")
	}
	tmpUint8 := d.TimerUnit<<5 | d.TimerValue
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (d *DLBufferingDuration) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	d.TimerUnit = uint8(data[idx]) >> 5 & Mask3
	d.TimerValue = uint8(data[idx]) & Mask5
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
