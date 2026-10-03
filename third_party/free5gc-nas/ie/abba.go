package ie

import "github.com/pkg/errors"

const (
	abbaValueMiniOctets int = 2
)

type ABBA struct {
	Abba []byte
}

func (i *ABBA) UnmarshalBinary(b []byte) error {
	if len(b) < abbaValueMiniOctets {
		return errors.Errorf("ABBA Unmarshal err: data length should be >= %d", abbaValueMiniOctets)
	}
	i.Abba = b
	return nil
}

func (i *ABBA) MarshalBinary() ([]byte, error) {
	if len(i.Abba) < abbaValueMiniOctets {
		return nil, errors.Errorf("ABBA Marshal err: data length should be >= %d", abbaValueMiniOctets)
	}
	return i.Abba, nil
}
