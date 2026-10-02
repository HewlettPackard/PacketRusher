package ie

import "github.com/pkg/errors"

const (
	AuthParamAUTNValueOctets int = 16
)

type AuthParamAUTN struct {
	Autn []byte
}

func (i *AuthParamAUTN) UnmarshalBinary(b []byte) error {
	if len(b) != AuthParamAUTNValueOctets {
		return errors.Errorf("AuthParamAUTN Unmarshal err: data length(%d) != %d", len(i.Autn), AuthParamAUTNValueOctets)
	}
	i.Autn = b
	return nil
}

func (i *AuthParamAUTN) MarshalBinary() ([]byte, error) {
	if len(i.Autn) != AuthParamAUTNValueOctets {
		return nil, errors.Errorf("AuthParamAUTN Marshal err: data length(%d) != %d", len(i.Autn), AuthParamAUTNValueOctets)
	}
	return i.Autn, nil
}
