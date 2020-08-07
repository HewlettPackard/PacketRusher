package pfcpType

import (
	"fmt"
)

type NodeReportType struct {
	Upfr bool
}

func (n *NodeReportType) MarshalBinary() (data []byte, err error) {
	// Octet 5
	tmpUint8 := btou(n.Upfr)
	data = append([]byte(""), byte(tmpUint8))

	return data, nil
}

func (n *NodeReportType) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	n.Upfr = utob(uint8(data[idx]) & BitMask1)
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
