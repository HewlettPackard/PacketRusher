package pfcpType

import (
	"fmt"
)

type MeasurementMethod struct {
	Event bool
	Volum bool
	Durat bool
}

func (m *MeasurementMethod) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if !m.Event && !m.Volum && !m.Durat {
		return []byte(""), fmt.Errorf("At least one of EVENT, VOLUM and DURAT shall be set")
	}
	tmpUint8 := btou(m.Event)<<2 |
		btou(m.Volum)<<1 |
		btou(m.Durat)
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (m *MeasurementMethod) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	m.Event = utob(uint8(data[idx]) & BitMask3)
	m.Volum = utob(uint8(data[idx]) & BitMask2)
	m.Durat = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	if !m.Event && !m.Volum && !m.Durat {
		return fmt.Errorf("None of EVENT, VOLUM and DURAT is set")
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
