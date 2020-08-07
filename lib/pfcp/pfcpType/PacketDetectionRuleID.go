package pfcpType

import (
	"encoding/binary"
	"fmt"
)

type PacketDetectionRuleID struct {
	RuleId uint16
}

func (p *PacketDetectionRuleID) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to 6
	data = make([]byte, 2)
	binary.BigEndian.PutUint16(data[idx:], p.RuleId)

	return data, nil
}

func (p *PacketDetectionRuleID) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 6
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	p.RuleId = binary.BigEndian.Uint16(data[idx:])
	idx = idx + 2

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
