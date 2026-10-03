package ie

import (
	"github.com/pkg/errors"
)

// RegResult5GS is detailed in 9.11.3.6 5GS registration result, 24.501
type RegResult5GS struct {
	// Name, uint8, Bits, Octet
	DisasterRoamingRegResult bool  // 7 -> 7 ,   3 -> 3
	EmergRegistered          bool  // 6 -> 6 ,   3 -> 3
	NSSAAPerformed           bool  // 5 -> 5 ,   3 -> 3
	SMSAllowed               bool  // 4 -> 4 ,   3 -> 3
	Value                    uint8 // 3 -> 1 ,   3 -> 3 , RegResult_3gpp, RegResult_Non3gpp, RegResult_3gppAndNon3gpp
}

const (
	RegResult_3gpp           uint8 = 0x01
	RegResult_Non3gpp        uint8 = 0x02
	RegResult_3gppAndNon3gpp uint8 = 0x03
)

const (
	// SMSAllowed
	SMSOverNASNotAllowed bool = false
	SMSOverNASAllowed    bool = true

	// NSSAAPerformed
	NwSliceSpecificAuthAndAuthorizationIsNotToBePerformed bool = false
	NwSliceSpecificAuthAndAuthorizationIsToBePerformed    bool = true

	// EmergRegistered
	NotRegisteredForEmergSvc bool = false
	RegisteredForEmergSvc    bool = true

	// DisasterRoamingRegResult
	DisasterRoamingNoAdditionalInfo bool = false
	// Request for registration for disaster roaming services accepted as
	// registration not for disaster roaming services
	DisasterRoamingAccepted_NotForDisasterRoaming bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RegResult5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The RegResult5GS  IE length(%d) is incorrect", len(b))
	}
	i.DisasterRoamingRegResult = GetBit7(b[0]) == 1
	i.EmergRegistered = GetBit6(b[0]) == 1
	i.NSSAAPerformed = GetBit5(b[0]) == 1
	i.SMSAllowed = GetBit4(b[0]) == 1
	i.Value = Get3Bits31(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RegResult5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit7(b[0], bool2uint8(i.DisasterRoamingRegResult))
	b[0] = SetBit6(b[0], bool2uint8(i.EmergRegistered))
	b[0] = SetBit5(b[0], bool2uint8(i.NSSAAPerformed))
	b[0] = SetBit4(b[0], bool2uint8(i.SMSAllowed))
	b[0] = Set3Bits31(b[0], i.Value)

	return b, nil
}
