package ie

import "github.com/pkg/errors"

// AllowedSSCMode implements the standard single-octet value; spare bits are ignored on decode.
type AllowedSSCMode struct {
	SSC3 uint8
	SSC2 uint8
	SSC1 uint8
}

func (i *AllowedSSCMode) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("AllowedSSCMode IE length %d, expected 1", len(b))
	}
	i.SSC3 = GetBit3(b[0])
	i.SSC2 = GetBit2(b[0])
	i.SSC1 = GetBit1(b[0])
	return nil
}

func (i *AllowedSSCMode) MarshalBinary() ([]byte, error) {
	if i.SSC3 > 1 || i.SSC2 > 1 || i.SSC1 > 1 {
		return nil, errors.Errorf("invalid allowed SSC mode bits")
	}
	return []byte{i.SSC3<<2 | i.SSC2<<1 | i.SSC1}, nil
}
