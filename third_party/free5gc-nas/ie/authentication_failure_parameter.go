package ie

import "github.com/pkg/errors"

// AuthFailureParam is detailed in 9.11.3.14 Authentication failure parameter, 24.501
// which references to 10.5.3.2.2 Authentication Failure parameter (UMTS and EPS authentication challenge)
type AuthFailureParam struct {
	Value []byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AuthFailureParam) UnmarshalBinary(b []byte) error {
	if len(b) != 14 {
		return errors.Errorf("Bad AuthFailureParam IE length(%d)", len(b))
	}
	i.Value = append(i.Value, b...)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AuthFailureParam) MarshalBinary() ([]byte, error) {
	ieLen := len(i.Value)
	if ieLen != 14 {
		return nil, errors.Errorf("Bad AuthFailureParam.Value length(%d)", ieLen)
	}
	b := make([]byte, ieLen)
	copy(b, i.Value)
	return b, nil
}
