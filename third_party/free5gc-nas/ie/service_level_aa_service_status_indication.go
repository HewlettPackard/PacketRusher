package ie

import "github.com/pkg/errors"

// SvcLvlAASvcStatusInd is detailed in 9.11.2.18 Service-level-AA service status indication, 24.501
type SvcLvlAASvcStatusInd struct {
	// Name, uint8, Bits, Octet
	uas bool // 1->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAASvcStatusInd) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("SvcLvlAASvcStatusInd bad IE len(%d)", len(b))
	}
	i.uas = (GetBit1(b[0]) == 1)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAASvcStatusInd) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit1(b[0], bool2uint8(i.uas))
	return b, nil
}
