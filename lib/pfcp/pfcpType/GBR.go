package pfcpType

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type GBR struct {
	UlGbr uint64 // 40-bit data
	DlGbr uint64 // 40-bit data
}

func (m *GBR) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 9
	if bits.Len64(m.UlGbr) > 40 {
		return []byte(""), fmt.Errorf("UL GBR shall not be greater than 40 bits binary integer")
	}
	tmpByteSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(tmpByteSlice, m.UlGbr)
	data = append([]byte(""), tmpByteSlice[3:]...)

	// Octet 10 to 14
	if bits.Len64(m.DlGbr) > 40 {
		return []byte(""), fmt.Errorf("DL GBR shall not be greater than 40 bits binary integer")
	}
	binary.BigEndian.PutUint64(tmpByteSlice, m.DlGbr)
	data = append(data, tmpByteSlice[3:]...)

	return data, nil
}

func (m *GBR) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 9
	if length < idx+5 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	tmpByteSlice := make([]byte, 8)
	copy(tmpByteSlice[3:], data[idx:idx+5])
	m.UlGbr = binary.BigEndian.Uint64(tmpByteSlice)
	idx = idx + 5

	// Octet 10 to 14
	if length < idx+5 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	copy(tmpByteSlice[3:], data[idx:idx+5])
	m.DlGbr = binary.BigEndian.Uint64(tmpByteSlice)
	idx = idx + 5

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
