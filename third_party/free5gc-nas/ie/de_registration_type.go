package ie

import "github.com/pkg/errors"

// DeregType is detailed in 9.11.3.20 De-registration type, 24.501
type DeregType struct {
	// Name, uint8, Bits, Octet
	Switchoff     bool  // 4 -> 4 ,   1 -> 1
	ReregRequired bool  // 3 -> 3 ,   1 -> 1
	AccessType    uint8 // 2 -> 1 ,   1 -> 1 AccessType_xxxx
}

const (
	AccessType_3gpp           uint8 = 0x01
	AccessType_Non3gpp        uint8 = 0x02
	AccessType_3gppAndNon3gpp uint8 = 0x03
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DeregType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The DeregType IE length(%d) is not correct", len(b))
	}
	i.Switchoff = GetBit4(b[0]) == 1
	i.ReregRequired = (GetBit3(b[0]) == 1)
	i.AccessType = Get2Bits21(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DeregType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit4(b[0], bool2uint8(i.Switchoff))
	b[0] = SetBit3(b[0], bool2uint8(i.ReregRequired))
	b[0] = Set2Bits21(b[0], i.AccessType)

	return b, nil
}
