package ie

import "github.com/pkg/errors"

// PDUSessId2 is detailed in 9.11.3.41 PDU session identity 2, 24.501
type PDUSessId2 struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 8 -> 1 ,   2 -> 2
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUSessId2) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The PDUSessId2 IE length(%d) is incorrect", len(b))
	}
	i.Value = b[0]
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUSessId2) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = i.Value

	return b, nil
}
