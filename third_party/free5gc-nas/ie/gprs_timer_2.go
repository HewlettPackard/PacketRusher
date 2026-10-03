package ie

import "github.com/pkg/errors"

// GPRS Timer 2 value is coded as octet 2 of the GPRS timer information element.
type GPRSTimer2 struct {
	Unit  GPRSTimerUnitType // 8 -> 6 ,   3 -> 3
	Value uint8             // 5 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *GPRSTimer2) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The GPRSTimer2 IE length(%d) is incorrect", len(b))
	}
	i.Unit = GPRSTimerUnitType(Get3Bits86(b[0]))
	i.Value = Get5Bits51(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *GPRSTimer2) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits86(b[0], uint8(i.Unit))
	b[0] = Set5Bits51(b[0], i.Value)
	return b, nil
}

func (i *GPRSTimer2) Set(sec uint32) {
	t := SecToGPRSTmr1(sec)
	i.Unit = GPRSTimerUnitType(Get3Bits86(t))
	i.Value = Get5Bits51(t)
}
