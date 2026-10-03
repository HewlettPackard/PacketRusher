package ie

import "github.com/pkg/errors"

// AlwaysonPDUSessReq is detailed in 9.11.4.4 Always-on PDU session requested, 24.501
type AlwaysonPDUSessReq struct {
	// Name, uint8, Bits, Octet
	APSR bool // 1->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AlwaysonPDUSessReq) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The AlwaysonPDUSessReq IE length(%d) is incorrect", len(b))
	}
	i.APSR = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AlwaysonPDUSessReq) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit1(b[0], bool2uint8(i.APSR))

	return b, nil
}
