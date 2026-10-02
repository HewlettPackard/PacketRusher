package ie

import (
	"github.com/pkg/errors"
)

type Psi struct {
	PSI [16]bool
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Psi) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return errors.Errorf("Bad Psi IE length(%d)", len(b))
	}
	i.PSI[7] = GetBit8(b[0]) == 1
	i.PSI[6] = GetBit7(b[0]) == 1
	i.PSI[5] = GetBit6(b[0]) == 1
	i.PSI[4] = GetBit5(b[0]) == 1
	i.PSI[3] = GetBit4(b[0]) == 1
	i.PSI[2] = GetBit3(b[0]) == 1
	i.PSI[1] = GetBit2(b[0]) == 1
	i.PSI[0] = GetBit1(b[0]) == 1
	i.PSI[15] = GetBit8(b[1]) == 1
	i.PSI[14] = GetBit7(b[1]) == 1
	i.PSI[13] = GetBit6(b[1]) == 1
	i.PSI[12] = GetBit5(b[1]) == 1
	i.PSI[11] = GetBit4(b[1]) == 1
	i.PSI[10] = GetBit3(b[1]) == 1
	i.PSI[9] = GetBit2(b[1]) == 1
	i.PSI[8] = GetBit1(b[1]) == 1

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Psi) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2)
	b[0] = SetBit8(b[0], bool2uint8(i.PSI[7]))
	b[0] = SetBit7(b[0], bool2uint8(i.PSI[6]))
	b[0] = SetBit6(b[0], bool2uint8(i.PSI[5]))
	b[0] = SetBit5(b[0], bool2uint8(i.PSI[4]))
	b[0] = SetBit4(b[0], bool2uint8(i.PSI[3]))
	b[0] = SetBit3(b[0], bool2uint8(i.PSI[2]))
	b[0] = SetBit2(b[0], bool2uint8(i.PSI[1]))
	b[0] = SetBit1(b[0], bool2uint8(i.PSI[0]))
	b[1] = SetBit8(b[1], bool2uint8(i.PSI[15]))
	b[1] = SetBit7(b[1], bool2uint8(i.PSI[14]))
	b[1] = SetBit6(b[1], bool2uint8(i.PSI[13]))
	b[1] = SetBit5(b[1], bool2uint8(i.PSI[12]))
	b[1] = SetBit4(b[1], bool2uint8(i.PSI[11]))
	b[1] = SetBit3(b[1], bool2uint8(i.PSI[10]))
	b[1] = SetBit2(b[1], bool2uint8(i.PSI[9]))
	b[1] = SetBit1(b[1], bool2uint8(i.PSI[8]))

	return b, nil
}
