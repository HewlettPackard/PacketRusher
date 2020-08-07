package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type URSEQN struct {
	UrseqnValue uint32
}

func (u *URSEQN) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to 8
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data[idx:], u.UrseqnValue)

	return data, nil
}

func (u *URSEQN) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 8
	if length < idx+4 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	u.UrseqnValue = binary.BigEndian.Uint32(data[idx:])
	idx = idx + 4

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
