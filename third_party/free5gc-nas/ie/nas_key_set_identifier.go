package ie

import (
	"github.com/pkg/errors"
)

type SecCtxType uint8

const (
	SecCtxTypeNative SecCtxType = 0x00
	SecCtxTypeMapped SecCtxType = 0x01
)

const (
	nasKeySetIdentifierValueOctets int = 1

	NASKeyNA       uint8 = 7 // No key is available, (UE -> Network)
	NASKeyReserved uint8 = 7 // (Network -> UE)
)

// NASKeySetId is detailed in 9.11.3.32 NAS key set identifier, 24.501
type NASKeySetId struct {
	Tsc SecCtxType
	Ksi uint8
}

func (i *NASKeySetId) UnmarshalBinary(b []byte) error {
	if len(b) != nasKeySetIdentifierValueOctets {
		return errors.Errorf("NASKeySetId Unmarshal err: data length is not enough")
	}
	i.Tsc = SecCtxType(GetBit4(b[0]))
	i.Ksi = Get3Bits31(b[0])
	return nil
}

func (i *NASKeySetId) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetBit4(b[0], uint8(i.Tsc))
	b[0] = Set3Bits31(b[0], i.Ksi)

	return b, nil
}
