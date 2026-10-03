package ie

import "github.com/pkg/errors"

// NwFeatureSupport5GSM is detailed in 9.11.4.18 5GSM network feature support, 24.501
type NwFeatureSupport5GSM struct {
	// Name, uint8, Bits, Octet
	EPTS1 uint8 // 1 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NwFeatureSupport5GSM) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("Bad NwFeatureSupport5GSM  IE length(%d)", len(b))
	}
	i.EPTS1 = GetBit1(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NwFeatureSupport5GSM) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit1(b[0], i.EPTS1)
	return b, nil
}
