package ie

import (
	"github.com/pkg/errors"
)

// RegType5GS is detailed in 9.11.3.7 5GS registration type, 24.501
type RegType5GS struct {
	// Name, uint8, Bits, Octet
	FOR_Pending bool  // 4 -> 4 ,   1 -> 1 , Follow On Req Pending
	Value       uint8 // 3 -> 1 ,   1 -> 1
}

const (
	// Value
	RegType_InitialReg                         uint8 = 0x01
	RegType_MobilityRegUpdating                uint8 = 0x02
	RegType_PeriodicRegUpdating                uint8 = 0x03
	RegType_EmergReg                           uint8 = 0x04
	RegType_SNPNOnboardingReg                  uint8 = 0x05
	RegType_DisasterRoamingMobilityRegUpdating uint8 = 0x06
	RegType_DisasterRoamingInitialReg          uint8 = 0x07

	// FOR
	ForType_NoFollowOnReqPending bool = false
	ForType_FollowOnReqPending   bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RegType5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The RegType5GS IE length(%d) is incorrect", len(b))
	}
	i.FOR_Pending = GetBit4(b[0]) == 1
	i.Value = Get3Bits31(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RegType5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit4(b[0], bool2uint8(i.FOR_Pending))
	b[0] = Set3Bits31(b[0], i.Value)

	return b, nil
}
