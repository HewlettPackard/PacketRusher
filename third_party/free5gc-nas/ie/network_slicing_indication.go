package ie

import "github.com/pkg/errors"

// NwSlicingInd is detailed in 9.11.3.36 Network slicing indication, 24.501
type NwSlicingInd struct {
	// Name, uint8, Bits, Octet
	DCNI  bool // 2 -> 2 ,   1 -> 1
	NSSCI bool // 1 -> 1 ,   1 -> 1
}

const (
	// DCNI
	IsNotCreatedFromDefaultNSSAI bool = false
	IsCreatedFromDefaultNSSAI    bool = true

	// NSSCI
	NwSliceSubscriptionNotChanged bool = false
	NwSliceSubscriptionChanged    bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NwSlicingInd) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("The NwSlicingInd IE length(%d) is not enough", len(b))
	}
	i.DCNI = GetBit2(b[0]) == 1
	i.NSSCI = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NwSlicingInd) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit2(b[0], bool2uint8(i.DCNI))
	b[0] = SetBit1(b[0], bool2uint8(i.NSSCI))

	return b, nil
}
