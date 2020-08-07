package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type SubsequentTimeThreshold struct {
	SubsequentTimeThreshold uint32
}

func (s *SubsequentTimeThreshold) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 8
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data, s.SubsequentTimeThreshold)

	return data, nil
}

func (s *SubsequentTimeThreshold) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 8
	if length < idx+4 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	s.SubsequentTimeThreshold = binary.BigEndian.Uint32(data[idx:])
	idx = idx + 4

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
