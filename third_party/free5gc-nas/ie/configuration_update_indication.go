package ie

import "github.com/pkg/errors"

// CfgUpdateInd is detailed in 9.11.3.18 Configuration update indication, 24.501
type CfgUpdateInd struct {
	// Name, uint8, Bits, Octet
	RED bool // 2->2,   1->1  Registration requested
	ACK bool // 1->1,   1->1  Acknowledgement
}

const (
	// RED
	RegistrationNotRequested bool = false
	RegistrationRequested    bool = true

	// ACK
	AckNotRequested bool = false
	AckRequested    bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CfgUpdateInd) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("Bad CfgUpdateInd IE length(%d)", len(b))
	}
	i.RED = GetBit2(b[0]) == 1
	i.ACK = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CfgUpdateInd) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit2(b[0], bool2uint8(i.RED))
	b[0] = SetBit1(b[0], bool2uint8(i.ACK))

	return b, nil
}
