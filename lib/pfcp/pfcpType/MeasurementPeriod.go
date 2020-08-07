package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type MeasurementPeriod struct {
	MeasurementPeriod uint32
}

func (m *MeasurementPeriod) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to 8
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data[idx:], m.MeasurementPeriod)

	return data, nil
}

func (m *MeasurementPeriod) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 8
	if length < idx+4 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	m.MeasurementPeriod = binary.BigEndian.Uint32(data[idx:])
	idx = idx + 4

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
