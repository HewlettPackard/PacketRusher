package pfcpType

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

type DLBufferingSuggestedPacketCount struct {
	PacketCountValue uint16
}

func (d *DLBufferingSuggestedPacketCount) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to (n+4)
	if bits.Len16(d.PacketCountValue) > 8 {
		data = make([]byte, 2)
		binary.BigEndian.PutUint16(data[idx:], d.PacketCountValue)
	} else {
		data = append(data, byte(d.PacketCountValue))
	}

	return data, nil
}

func (d *DLBufferingSuggestedPacketCount) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to (n+4)
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	if length == 1 {
		d.PacketCountValue = uint16(data[idx])
		idx = idx + 1
	} else {
		d.PacketCountValue = binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2
	}

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
