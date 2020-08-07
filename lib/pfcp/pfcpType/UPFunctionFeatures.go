package pfcpType

import (
	"encoding/binary"
	"fmt"
)

const (
	UpFunctionFeaturesBucp  uint16 = 1
	UpFunctionFeaturesDdnd  uint16 = 1 << 1
	UpFunctionFeaturesDlbd  uint16 = 1 << 2
	UpFunctionFeaturesTrst  uint16 = 1 << 3
	UpFunctionFeaturesFtup  uint16 = 1 << 4
	UpFunctionFeaturesPfdm  uint16 = 1 << 5
	UpFunctionFeaturesHeeu  uint16 = 1 << 6
	UpFunctionFeaturesTreu  uint16 = 1 << 7
	UpFunctionFeaturesEmpu  uint16 = 1 << 8
	UpFunctionFeaturesPdiu  uint16 = 1 << 9
	UpFunctionFeaturesUdbc  uint16 = 1 << 10
	UpFunctionFeaturesQuoac uint16 = 1 << 11
	UpFunctionFeaturesTrace uint16 = 1 << 12
	UpFunctionFeaturesFrrt  uint16 = 1 << 13
)

type UPFunctionFeatures struct {
	SupportedFeatures uint16
}

func (u *UPFunctionFeatures) MarshalBinary() (data []byte, err error) {
	var idx uint16 = 0
	// Octet 5 to 6
	data = make([]byte, 2)
	binary.LittleEndian.PutUint16(data[idx:], u.SupportedFeatures)

	return data, nil
}

func (u *UPFunctionFeatures) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 6
	if length < idx+2 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	u.SupportedFeatures = binary.LittleEndian.Uint16(data[idx:])
	idx = idx + 2

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
