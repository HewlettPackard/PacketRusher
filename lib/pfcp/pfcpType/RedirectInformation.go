package pfcpType

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type RedirectInformation struct {
	RedirectAddressType         uint8 // 0x00001111
	RedirectServerAddressLength uint16
	RedirectServerAddress       []byte
}

func (r *RedirectInformation) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5
	if bits.Len8(r.RedirectAddressType) > 4 {
		return []byte(""), fmt.Errorf("Redirect address type shall not be greater than 4 bits binary integer")
	}
	data = append([]byte(""), byte(r.RedirectAddressType))
	idx = idx + 1

	// Octet 6 to 7
	data = append(data, make([]byte, 2)...)
	binary.BigEndian.PutUint16(data[idx:], r.RedirectServerAddressLength)

	// Octet 8 to (8+a)
	if len(r.RedirectServerAddress) != int(r.RedirectServerAddressLength) {
		return []byte(""), fmt.Errorf("Unmatch length of redirect server address: Expect %d, got %d", r.RedirectServerAddressLength, len(r.RedirectServerAddress))
	}
	data = append(data, r.RedirectServerAddress...)

	return data, nil
}

func (r *RedirectInformation) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	r.RedirectAddressType = uint8(data[idx]) & Mask4
	idx = idx + 1

	// Octet 6 to 7
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	r.RedirectServerAddressLength = binary.BigEndian.Uint16(data[idx:])
	idx = idx + 2

	// Octet 8 to (8+a)
	if length < idx+r.RedirectServerAddressLength {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	r.RedirectServerAddress = data[idx : idx+r.RedirectServerAddressLength]
	idx = idx + r.RedirectServerAddressLength

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
