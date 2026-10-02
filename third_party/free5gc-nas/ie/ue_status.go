package ie

import "github.com/pkg/errors"

// UEStatus is detailed in 9.11.3.56 UE status, 24.501
type UEStatus struct {
	// Name, uint8, Bits, Octet
	N1ModeReg bool // 2 -> 2 ,   3 -> 3
	S1ModeReg bool // 1 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UEStatus) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("UEStatus bad IE len(%d)", len(b))
	}
	i.N1ModeReg = GetBit2(b[0]) == 1
	i.S1ModeReg = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UEStatus) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit2(b[0], bool2uint8(i.N1ModeReg))
	b[0] = SetBit1(b[0], bool2uint8(i.S1ModeReg))

	return b, nil
}
