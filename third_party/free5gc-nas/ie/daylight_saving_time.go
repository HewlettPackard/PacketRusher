package ie

import "github.com/pkg/errors"

// 10.5.3.12 Daylight Saving Time, 24.008
type DST struct {
	Value uint8 // 2 -> 1 ,   3 -> 3
}

type TypeDST uint8

const (
	NoAdjustment     TypeDST = 0x00
	HourAdjustment_1 TypeDST = 0x01
	HourAdjustment_2 TypeDST = 0x02
	Reserved         TypeDST = 0x03
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DST) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("Bad DST IE length(%d)", len(b))
	}

	i.Value = Get2Bits21(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DST) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set2Bits21(b[0], i.Value)

	return b, nil
}
