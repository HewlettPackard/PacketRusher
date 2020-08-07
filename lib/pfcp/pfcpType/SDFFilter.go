package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type SDFFilter struct {
	Bid                     bool
	Fl                      bool
	Spi                     bool
	Ttc                     bool
	Fd                      bool
	LengthOfFlowDescription uint16
	FlowDescription         []byte
	TosTrafficClass         []byte
	SecurityParameterIndex  []byte
	FlowLabel               []byte
	SdfFilterId             uint32
}

func (s *SDFFilter) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5
	tmpUint8 := btou(s.Bid)<<4 |
		btou(s.Fl)<<3 |
		btou(s.Spi)<<2 |
		btou(s.Ttc)<<1 |
		btou(s.Fd)
	data = append([]byte(""), byte(tmpUint8))
	idx = idx + 1

	// Octet 6 (spare)
	data = append(data, byte(0))
	idx = idx + 1

	// Octet m to (m+1)
	// Octet (m+2) to p
	if s.Fd {
		if s.LengthOfFlowDescription == 0 {
			return []byte(""), fmt.Errorf("Length of flow description shall be present if FD is set")
		}
		data = append(data, make([]byte, 2)...)
		binary.BigEndian.PutUint16(data[idx:], s.LengthOfFlowDescription)
		idx = idx + 2

		if len(s.FlowDescription) != int(s.LengthOfFlowDescription) {
			return []byte(""), fmt.Errorf("Unmatch length of flow description: Expect %d, got %d", s.LengthOfFlowDescription, len(s.FlowDescription))
		}
		data = append(data, s.FlowDescription...)
		idx = idx + uint16(len(s.FlowDescription))
	}

	// Octet s to (s+1)
	if s.Ttc {
		if len(s.TosTrafficClass) != 2 {
			return []byte(""), fmt.Errorf("ToS traffic class shall be exactly two bytes")
		}
		data = append(data, s.TosTrafficClass...)
		idx = idx + 2
	}

	// Octet t to (t+3)
	if s.Spi {
		if len(s.SecurityParameterIndex) != 4 {
			return []byte(""), fmt.Errorf("Security parameter index shall be exactly four bytes")
		}
		data = append(data, s.SecurityParameterIndex...)
		idx = idx + 4
	}

	// Octet v to (v+2)
	if s.Fl {
		if len(s.FlowLabel) != 3 {
			return []byte(""), fmt.Errorf("Flow label shall be exactly three bytes")
		}
		data = append(data, s.FlowLabel...)
		idx = idx + 3
	}

	// Octet w to (w+3)
	if s.Bid {
		data = append(data, make([]byte, 4)...)
		binary.BigEndian.PutUint32(data[idx:], s.SdfFilterId)
	}

	return data, nil
}

func (s *SDFFilter) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	s.Bid = utob(uint8(data[idx]) & BitMask5)
	s.Fl = utob(uint8(data[idx]) & BitMask4)
	s.Spi = utob(uint8(data[idx]) & BitMask3)
	s.Ttc = utob(uint8(data[idx]) & BitMask2)
	s.Fd = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet 6 (spare)
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	idx = idx + 1

	// Octet m to (m+1)
	// Octet (m+2) to p
	if s.Fd {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.LengthOfFlowDescription = binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2

		if length < idx+s.LengthOfFlowDescription {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.FlowDescription = data[idx : idx+s.LengthOfFlowDescription]
		idx = idx + s.LengthOfFlowDescription
	}

	// Octet s to (s+1)
	if s.Ttc {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.TosTrafficClass = data[idx : idx+2]
		idx = idx + 2
	}

	// Octet t to (t+3)
	if s.Spi {
		if length < idx+4 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.SecurityParameterIndex = data[idx : idx+4]
		idx = idx + 4
	}

	// Octet v to (v+2)
	if s.Fl {
		if length < idx+3 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.FlowLabel = data[idx : idx+3]
		idx = idx + 3
	}

	// Octet w to (w+3)
	if s.Bid {
		if length < idx+4 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.SdfFilterId = binary.BigEndian.Uint32(data[idx : idx+4])
		idx = idx + 4
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
