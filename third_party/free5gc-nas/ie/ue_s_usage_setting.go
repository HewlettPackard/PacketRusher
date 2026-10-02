package ie

import "github.com/pkg/errors"

// UesUsageSetting is detailed in 9.11.3.55 UE's usage setting, 24.501
type UesUsageSetting struct {
	// Name, uint8, Bits, Octet
	UESUsageSetting uint8 // 1 -> 1 ,   3 -> 3
}

type TypeUesUsage uint8

const (
	UsageType_VoiceCentric TypeUesUsage = 0x00
	UsageType_DataCentric  TypeUesUsage = 0x01
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UesUsageSetting) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("UesUsageSetting bad IE len(%d)", len(b))
	}
	i.UESUsageSetting = GetBit1(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UesUsageSetting) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit1(b[0], i.UESUsageSetting)

	return b, nil
}
