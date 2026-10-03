package ie

import "github.com/pkg/errors"

// AuthRspParam is detailed in 9.11.3.17 Authentication response parameter, 24.501
type AuthRspParam struct {
	Res []byte // 8 -> 1 , 3 -> 18
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AuthRspParam) UnmarshalBinary(b []byte) error {
	if len(b) < 4 {
		return errors.Errorf("Bad AuthRspParam IE length(%d)", len(b))
	}
	i.Res = append(i.Res, b...)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AuthRspParam) MarshalBinary() ([]byte, error) {
	ieLen := len(i.Res)
	if ieLen < 4 {
		return nil, errors.Errorf("Bad AuthRspParam.Res length(%d)", ieLen)
	}
	b := make([]byte, ieLen)
	copy(b, i.Res)
	return b, nil
}
