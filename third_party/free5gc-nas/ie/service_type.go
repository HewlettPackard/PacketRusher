package ie

import (
	"github.com/pkg/errors"
)

// SvcType is detailed in 9.11.3.50 Service type, 24.501
type SvcType struct {
	// Name, uint8, Bits, Octet
	Value ConstSvcType // 4 -> 1 ,   1 -> 1
}

type ConstSvcType uint8

const (
	SvcType_Signalling         ConstSvcType = 0x00
	SvcType_Data               ConstSvcType = 0x01
	SvcType_MobileTermSvc      ConstSvcType = 0x02
	SvcType_EmergSvc           ConstSvcType = 0x03
	SvcType_EmergSvcFlbk       ConstSvcType = 0x04
	SvcType_HighPriorityAccess ConstSvcType = 0x05
	SvcType_ElevatedSignalling ConstSvcType = 0x06
	// 0x07 ~ 0x0B are unused
	SvcType_Reserved01 ConstSvcType = 0x0C
	SvcType_Reserved02 ConstSvcType = 0x0D
	SvcType_Reserved03 ConstSvcType = 0x0E
	SvcType_Reserved04 ConstSvcType = 0x0F
)

var svcTypeMap = map[ConstSvcType]string{
	SvcType_Signalling:         "Signaling",
	SvcType_Data:               "Data",
	SvcType_MobileTermSvc:      "MobileTermSvc",
	SvcType_EmergSvc:           "EmergSvc",
	SvcType_EmergSvcFlbk:       "EmergSvcFlbk",
	SvcType_HighPriorityAccess: "HighPriorityAccess",
	SvcType_ElevatedSignalling: "ElevatedSignalling",
	SvcType_Reserved01:         "Reserved01 ",
	SvcType_Reserved02:         "Reserved02 ",
	SvcType_Reserved03:         "Reserved03 ",
	SvcType_Reserved04:         "Reserved04 ",
}

func (i *SvcType) String() string {
	if i == nil {
		return ""
	}
	return svcTypeMap[i.Value]
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("svcType bad IE len(%d)", len(b))
	}
	switch Get4Bits41(b[0]) {
	case 0x07:
		fallthrough
	case 0x08:
		i.Value = SvcType_Signalling
	case 0x09:
		fallthrough
	case 0x0A:
		fallthrough
	case 0x0B:
		i.Value = SvcType_Data
	default:
		i.Value = ConstSvcType(Get4Bits41(b[0]))
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set4Bits41(b[0], uint8(i.Value))

	return b, nil
}
