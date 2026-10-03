package ie

import "github.com/pkg/errors"

// SMSInd implements the standard single-octet value; spare bits are ignored on decode.
type SMSInd struct {
	SAI bool
}

func (i *SMSInd) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("SMSInd IE length %d, expected 1", len(b))
	}
	i.SAI = GetBit1(b[0]) == 1
	return nil
}

func (i *SMSInd) MarshalBinary() ([]byte, error) {

	return []byte{bool2uint8(i.SAI)}, nil
}
