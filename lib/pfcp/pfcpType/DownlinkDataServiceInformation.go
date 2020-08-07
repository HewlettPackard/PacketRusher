package pfcpType

import (
	"fmt"
	"math/bits"
)

type DownlinkDataServiceInformation struct {
	Qfii                        bool
	Ppi                         bool
	PagingPolicyIndicationValue uint8 // 0x00111111
	Qfi                         uint8 // 0x00111111
}

func (d *DownlinkDataServiceInformation) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(d.Qfii)<<1 | btou(d.Ppi)
	data = append([]byte(""), byte(tmpUint8))

	// Octet m
	if d.Ppi {
		if bits.Len8(d.PagingPolicyIndicationValue) > 6 {
			return []byte(""), fmt.Errorf("Paging policy information data shall not be greater than 6 bits binary integer")
		}
		data = append(data, byte(d.PagingPolicyIndicationValue))
	}

	// Octet p
	if d.Qfii {
		if bits.Len8(d.Qfi) > 6 {
			return []byte(""), fmt.Errorf("QFI shall not be greater than 6 bits binary integer")
		}
		data = append(data, byte(d.Qfi))
	}

	return data, nil
}

func (d *DownlinkDataServiceInformation) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	d.Qfii = utob(uint8(data[idx]) & BitMask2)
	d.Ppi = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet m
	if d.Ppi {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		d.PagingPolicyIndicationValue = uint8(data[idx]) & Mask6
		idx = idx + 1
	}

	// Octet p
	if d.Qfii {
		if length < idx+1 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		d.Qfi = uint8(data[idx]) & Mask6
		idx = idx + 1
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
