package pfcpType

import (
	"fmt"
)

type ReportType struct {
	Upir bool
	Erir bool
	Usar bool
	Dldr bool
}

func (r *ReportType) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(r.Upir)<<3 |
		btou(r.Erir)<<2 |
		btou(r.Usar)<<1 |
		btou(r.Dldr)
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (r *ReportType) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	r.Upir = utob(uint8(data[idx]) & BitMask4)
	r.Erir = utob(uint8(data[idx]) & BitMask3)
	r.Usar = utob(uint8(data[idx]) & BitMask2)
	r.Dldr = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
