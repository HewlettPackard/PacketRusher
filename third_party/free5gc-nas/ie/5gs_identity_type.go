package ie

import "github.com/pkg/errors"

const (
	IdType_5GS_None       uint8 = 0x00
	IdType_5GS_SUCI       uint8 = 0x01
	IdType_5GS_GUTI       uint8 = 0x02
	IdType_5GS_IMEI       uint8 = 0x03
	IdType_5GS_TMSI       uint8 = 0x04
	IdType_5GS_IMEISV     uint8 = 0x05
	IdType_5GS_MACAddress uint8 = 0x06
	IdType_5GS_EUI64      uint8 = 0x07
)

const (
	idType5GSValueOctets int = 1
)

type IdType5GS struct {
	IdType uint8
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *IdType5GS) UnmarshalBinary(b []byte) error {
	if len(b) != idType5GSValueOctets {
		return errors.Errorf("IdType5GS Unmarshal err: data length is not enough")
	}
	if i.IdType = Get3Bits31(b[0]); i.IdType == IdType_5GS_None {
		// All other values are unused and shall be interpreted as "SUCI", if received by the UE.
		i.IdType = IdType_5GS_SUCI
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *IdType5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits31(b[0], i.IdType)

	return b, nil
}
