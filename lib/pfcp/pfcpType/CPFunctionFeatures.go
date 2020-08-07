package pfcpType

import (
	"fmt"
)

const (
	CpFunctionFeaturesLoad uint8 = 1
	CpFunctionFeaturesOvrl uint8 = 1 << 1
)

type CPFunctionFeatures struct {
	SupportedFeatures uint8
}

func (c *CPFunctionFeatures) MarshalBinary() (data []byte, err error) {
	// Octet 5
	data = append([]byte(""), byte(c.SupportedFeatures))

	return data, nil
}

func (c *CPFunctionFeatures) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	c.SupportedFeatures = uint8(data[idx])
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
