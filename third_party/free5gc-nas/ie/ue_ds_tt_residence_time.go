package ie

import "github.com/pkg/errors"

// UEDSTTResidenceTime is detailed in 9.11.4.26 UE-DS-TT residence time, 24.501
type UEDSTTResidenceTime struct {
	// Name, uint8, Bits, Octet
	ResidenceTime [8]byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UEDSTTResidenceTime) UnmarshalBinary(b []byte) error {
	if len(b) != 8 {
		return errors.Errorf("Bad UEDSTTResidenceTime IE length(%d)", len(b))
	}
	copy(i.ResidenceTime[:], b)

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UEDSTTResidenceTime) MarshalBinary() ([]byte, error) {
	b := make([]byte, 8)
	copy(b, i.ResidenceTime[:])

	return b, nil
}
