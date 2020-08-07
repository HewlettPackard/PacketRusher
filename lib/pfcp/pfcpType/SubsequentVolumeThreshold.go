package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type SubsequentVolumeThreshold struct {
	Dlvol          bool
	Ulvol          bool
	Tovol          bool
	TotalVolume    uint64
	UplinkVolume   uint64
	DownlinkVolume uint64
}

func (s *SubsequentVolumeThreshold) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5
	tmpUint8 := btou(s.Dlvol)<<2 |
		btou(s.Ulvol)<<1 |
		btou(s.Tovol)
	data = append([]byte(""), byte(tmpUint8))
	idx = idx + 1

	// Octet m to (m+7)
	if s.Tovol {
		data = append(data, make([]byte, 8)...)
		binary.BigEndian.PutUint64(data[idx:], s.TotalVolume)
		idx = idx + 8
	}

	// Octet p to (p+7)
	if s.Ulvol {
		data = append(data, make([]byte, 8)...)
		binary.BigEndian.PutUint64(data[idx:], s.UplinkVolume)
		idx = idx + 8
	}

	// Octet q to (q+7)
	if s.Dlvol {
		data = append(data, make([]byte, 8)...)
		binary.BigEndian.PutUint64(data[idx:], s.DownlinkVolume)
	}

	return data, nil
}

func (s *SubsequentVolumeThreshold) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	s.Dlvol = utob(uint8(data[idx]) & BitMask3)
	s.Ulvol = utob(uint8(data[idx]) & BitMask2)
	s.Tovol = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	// Octet m to (m+7)
	if s.Tovol {
		if length < idx+8 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.TotalVolume = binary.BigEndian.Uint64(data[idx:])
		idx = idx + 8
	}

	// Octet p to (p+7)
	if s.Ulvol {
		if length < idx+8 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.UplinkVolume = binary.BigEndian.Uint64(data[idx:])
		idx = idx + 8
	}

	// Octet q to (q+7)
	if s.Dlvol {
		if length < idx+8 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		s.DownlinkVolume = binary.BigEndian.Uint64(data[idx:])
		idx = idx + 8
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
