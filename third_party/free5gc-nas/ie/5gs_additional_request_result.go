package ie

import "github.com/pkg/errors"

const (
	PagingRestriction_NoInfo   = 0x00
	PagingRestriction_Accepted = 0x01
	PagingRestriction_Rejected = 0x02
)

// AdditionalReqResult5GS is detailed in 9.11.3.81 5GS additional request result, 24.501
type AdditionalReqResult5GS struct {
	// Name, uint8, Bits, Octet
	PRD uint8 // 2 -> 1 ,   3 -> 3 // Paging Restriction Decision
}

func (i *AdditionalReqResult5GS) inRange() bool {
	switch i.PRD {
	case PagingRestriction_NoInfo, PagingRestriction_Accepted, PagingRestriction_Rejected:
		return true
	}
	return false
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AdditionalReqResult5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("AdditionalReqResult5GS bad IE len(%d)", len(b))
	}
	i.PRD = Get2Bits21(b[0])
	if !i.inRange() {
		return errors.Errorf("AdditionalReqResult5GS bad value %d", i.PRD)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AdditionalReqResult5GS) MarshalBinary() ([]byte, error) {
	if !i.inRange() {
		return nil, errors.Errorf("AdditionalReqResult5GS bad value %d", i.PRD)
	}
	b := make([]byte, 1)
	b[0] = Set2Bits21(b[0], i.PRD)
	return b, nil
}
