package pfcpType

import (
	"fmt"
	"math/bits"
)

type HeaderEnrichment struct {
	HeaderType               uint8 // 0x00011111
	LengthOfHeaderFieldName  uint8
	HeaderFieldName          []byte
	LengthOfHeaderFieldValue uint8
	HeaderFieldValue         []byte
}

func (h *HeaderEnrichment) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(h.HeaderType) > 5 {
		return []byte(""), fmt.Errorf("Header type shall not be greater than 5 bits binary integer")
	}
	data = append([]byte(""), byte(h.HeaderType))

	// Octet 6
	data = append(data, byte(h.LengthOfHeaderFieldName))

	// Octet 7 to m
	if len(h.HeaderFieldName) != int(h.LengthOfHeaderFieldName) {
		return []byte(""), fmt.Errorf("Unmatch length of header field name: Expect %d, got %d", h.LengthOfHeaderFieldName, len(h.HeaderFieldName))
	}
	data = append(data, h.HeaderFieldName...)

	// Octet p
	data = append(data, byte(h.LengthOfHeaderFieldValue))

	// Octet (p+1) to q
	if len(h.HeaderFieldValue) != int(h.LengthOfHeaderFieldValue) {
		return []byte(""), fmt.Errorf("Unmatch length of header field name: Expect %d, got %d", h.LengthOfHeaderFieldValue, len(h.HeaderFieldValue))
	}
	data = append(data, h.HeaderFieldValue...)

	return data, nil
}

func (h *HeaderEnrichment) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	h.HeaderType = uint8(data[idx]) & Mask5
	idx = idx + 1

	// Octet 6
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	h.LengthOfHeaderFieldName = data[idx]
	idx = idx + 1

	// Octet 7 to m
	if length < idx+uint16(h.LengthOfHeaderFieldName) {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	h.HeaderFieldName = data[idx : idx+uint16(h.LengthOfHeaderFieldName)]
	idx = idx + uint16(h.LengthOfHeaderFieldName)

	// Octet p
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	h.LengthOfHeaderFieldValue = data[idx]
	idx = idx + 1

	// Octet (p+1) to q
	if length < idx+uint16(h.LengthOfHeaderFieldValue) {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	h.HeaderFieldValue = data[idx : idx+uint16(h.LengthOfHeaderFieldValue)]
	idx = idx + uint16(h.LengthOfHeaderFieldValue)

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
