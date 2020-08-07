package pfcpType

import (
	"fmt"
	"math/bits"
)

const (
	DestinationInterfaceAccess uint8 = iota
	DestinationInterfaceCore
	DestinationInterfaceSgiLanN6Lan
	DestinationInterfaceCpFunction
	DestinationInterfaceLiFunction
)

type DestinationInterface struct {
	InterfaceValue uint8 // 0x00001111
}

func (d *DestinationInterface) MarshalBinary() (data []byte, err error) {
	// Octet 5
	if bits.Len8(d.InterfaceValue) > 4 {
		return []byte(""), fmt.Errorf("Interface data shall not be greater than 4 bits binary integer")
	}
	data = append([]byte(""), byte(d.InterfaceValue))

	return data, nil
}

func (d *DestinationInterface) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	d.InterfaceValue = uint8(data[idx]) & Mask4
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
