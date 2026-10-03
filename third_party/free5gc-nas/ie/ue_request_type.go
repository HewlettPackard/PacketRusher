package ie

import "github.com/pkg/errors"

const (
	UEReqType_Reserve        uint8 = 0
	UEReqType_NasConnRelease uint8 = 1
	UEReqType_PagingRej      uint8 = 2
)

// UEReqType is detailed in 9.11.3.76 UE request type, 24.501 (9.9.3.65 in 3GPP TS 24.301)
type UEReqType struct {
	// Name, uint8, Bits, Octet
	ReqType uint8 // 4 -> 1 ,   3 -> 3
}

func (i *UEReqType) inRange() bool {
	switch i.ReqType {
	case UEReqType_Reserve, UEReqType_NasConnRelease, UEReqType_PagingRej:
		return true
	}
	return false
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UEReqType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("UEReqType bad IE len(%d)", len(b))
	}
	i.ReqType = Get4Bits41(b[0])
	if !i.inRange() {
		return errors.Errorf("UEReqType bad value %d", i.ReqType)
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UEReqType) MarshalBinary() ([]byte, error) {
	if !i.inRange() {
		return nil, errors.Errorf("UEReqType bad value %d", i.ReqType)
	}
	b := make([]byte, 1)
	b[0] = Set4Bits41(b[0], i.ReqType)
	return b, nil
}
