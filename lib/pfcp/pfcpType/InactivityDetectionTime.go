package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type InactivityDetectionTime struct {
	InactivityDetectionTime uint32
}

func (i *InactivityDetectionTime) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 8
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data, i.InactivityDetectionTime)

	return data, nil
}

func (i *InactivityDetectionTime) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 8
	if length < idx+4 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	i.InactivityDetectionTime = binary.BigEndian.Uint32(data[idx:])
	idx = idx + 4

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
