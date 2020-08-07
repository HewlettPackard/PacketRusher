package pfcpType

import (
	"fmt"
)

const (
	OuterHeaderRemovalGtpUUdpIpv4 uint8 = iota
	OuterHeaderRemovalGtpUUdpIpv6
	OuterHeaderRemovalUdpIpv4
	OuterHeaderRemovalUdpIpv6
)

type OuterHeaderRemoval struct {
	OuterHeaderRemovalDescription uint8
}

func (o *OuterHeaderRemoval) MarshalBinary() (data []byte, err error) {
	// Octet 5
	data = append([]byte(""), byte(o.OuterHeaderRemovalDescription))

	return data, nil
}

func (o *OuterHeaderRemoval) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	o.OuterHeaderRemovalDescription = uint8(data[idx])
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
