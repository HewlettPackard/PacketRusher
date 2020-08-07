package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type OffendingIE struct {
	TypeOfOffendingIe uint16
}

func (o *OffendingIE) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to 6
	data = make([]byte, 2)
	binary.BigEndian.PutUint16(data[idx:], o.TypeOfOffendingIe)

	return data, nil
}

func (o *OffendingIE) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 6
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	o.TypeOfOffendingIe = binary.BigEndian.Uint16(data[idx:])
	idx = idx + 2

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
