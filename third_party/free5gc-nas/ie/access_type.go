package ie

import "github.com/pkg/errors"

// AccessType is detailed in 9.11.2.1A Access type, 24.501
type AccessType struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 2->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AccessType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("AccessType bad length(%d)", len(b))
	}
	i.Value = Get2Bits21(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AccessType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set2Bits21(b[0], i.Value)
	return b, nil
}
