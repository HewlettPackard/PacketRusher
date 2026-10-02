package ie

import "github.com/pkg/errors"

// PayloadCntrType is detailed in 9.11.3.40 Payload container type, 24.501
type PayloadCntrType struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 4 -> 1 ,   1 -> 1
}

const (
	PayloadCntrType_N1SMInfo         uint8 = 0x01
	PayloadCntrType_SMS              uint8 = 0x02
	PayloadCntrType_LPP              uint8 = 0x03
	PayloadCntrType_SOR              uint8 = 0x04
	PayloadCntrType_UEPolicy         uint8 = 0x05
	PayloadCntrType_UEParamsUpdate   uint8 = 0x06
	PayloadCntrType_LocationSvcMsg   uint8 = 0x07
	PayloadCntrType_CiotUserData     uint8 = 0x08
	PayloadCntrType_SvcLvlAACntr     uint8 = 0x09
	PayloadCntrType_EventNotif       uint8 = 0x0a
	PayloadCntrType_MultiplePayloads uint8 = 0x0f
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PayloadCntrType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The PayloadCntrType  IE length(%d) is incorrect", len(b))
	}
	i.Value = Get4Bits41(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PayloadCntrType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set4Bits41(b[0], i.Value)

	return b, nil
}
