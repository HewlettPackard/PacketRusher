package pfcpType

import (
	"fmt"
)

type TransportLevelMarking struct {
	TosTrafficClass []byte
}

func (t *TransportLevelMarking) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 6
	if len(t.TosTrafficClass) != 2 {
		return []byte(""), fmt.Errorf("ToS/Traffic class shall be exactly two bytes")
	}
	data = t.TosTrafficClass

	return data, nil
}

func (t *TransportLevelMarking) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 6
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	t.TosTrafficClass = data[idx : idx+2]
	idx = idx + 2

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
