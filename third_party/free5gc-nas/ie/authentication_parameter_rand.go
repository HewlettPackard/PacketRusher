package ie

import "github.com/pkg/errors"

const (
	AuthParamRANDValueOctets int = 16
)

type AuthParamRAND struct {
	Rand []byte
}

func (i *AuthParamRAND) UnmarshalBinary(b []byte) error {
	if len(b) != AuthParamRANDValueOctets {
		return errors.Errorf("AuthParamRAND Unmarshal err: data length(%d) != %d", len(b), AuthParamRANDValueOctets)
	}
	i.Rand = b
	return nil
}

func (i *AuthParamRAND) MarshalBinary() ([]byte, error) {
	if len(i.Rand) != AuthParamRANDValueOctets {
		return nil, errors.Errorf("AuthParamRAND Marshal err: data length(%d) != %d", len(i.Rand), AuthParamRANDValueOctets)
	}
	return i.Rand, nil
}
