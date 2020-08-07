package pfcpType

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type MBR struct {
	UlMbr uint64 // 40-bit data
	DlMbr uint64 // 40-bit data
}

func (m *MBR) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 9
	if bits.Len64(m.UlMbr) > 40 {
		return []byte(""), fmt.Errorf("UL GBR shall not be greater than 40 bits binary integer")
	}
	tmpByteSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpByteSlice, m.UlMbr)
	data = append([]byte(""), tmpByteSlice[3:]...)

	// Octet 10 to 14
	if bits.Len64(m.DlMbr) > 40 {
		return []byte(""), fmt.Errorf("DL GBR shall not be greater than 40 bits binary integer")
	}
	binary.BigEndian.PutUint64(tmpByteSlice, m.DlMbr)
	data = append(data, tmpByteSlice[3:]...)

	return data, nil
}

func (m *MBR) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 9
	if length < idx+5 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	tmpByteSlice := make([]byte, 8)
	copy(tmpByteSlice[3:], data[idx:idx+5])
	m.UlMbr = binary.BigEndian.Uint64(tmpByteSlice)
	idx = idx + 5

	// Octet 10 to 14
	if length < idx+5 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	copy(tmpByteSlice[3:], data[idx:idx+5])
	m.DlMbr = binary.BigEndian.Uint64(tmpByteSlice)
	idx = idx + 5

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
