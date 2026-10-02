package ie

import (
	"github.com/pkg/errors"
)

type (
	AlgIntegrity uint8
	AlgCiphering uint8
)

// NASSecAlgos is detailed in 9.11.3.34 NAS security algorithms, 24.501
type NASSecAlgos struct {
	// Name, uint8, Bits, Octet
	CipheringAlgo AlgCiphering // 8 -> 5 ,   2 -> 2
	MsgIntAlgo    AlgIntegrity // 4 -> 1 ,   2 -> 2
}

const (
	EncAlgo_5GEA0    AlgCiphering = 0x00
	EncAlgo_1285GEA1 AlgCiphering = 0x01
	EncAlgo_1285GEA2 AlgCiphering = 0x02
	EncAlgo_1285GEA3 AlgCiphering = 0x03
	EncAlgo_5GEA4    AlgCiphering = 0x04
	EncAlgo_5GEA5    AlgCiphering = 0x05
	EncAlgo_5GEA6    AlgCiphering = 0x06
	EncAlgo_5GEA7    AlgCiphering = 0x07

	IntegrityAlgo_5GIA0    AlgIntegrity = 0x00
	IntegrityAlgo_1285GIA1 AlgIntegrity = 0x01
	IntegrityAlgo_1285GIA2 AlgIntegrity = 0x02
	IntegrityAlgo_1285GIA3 AlgIntegrity = 0x03
	IntegrityAlgo_5GIA4    AlgIntegrity = 0x04
	IntegrityAlgo_5GIA5    AlgIntegrity = 0x05
	IntegrityAlgo_5GIA6    AlgIntegrity = 0x06
	IntegrityAlgo_5GIA7    AlgIntegrity = 0x07
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NASSecAlgos) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The NASSecAlgos IE length(%d) is incorrect", len(b))
	}

	tmp85, tmp41 := GetHalfValue(b[0])
	i.CipheringAlgo = AlgCiphering(tmp85)
	i.MsgIntAlgo = AlgIntegrity(tmp41)

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NASSecAlgos) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = SetHalfValue(uint8(i.CipheringAlgo), uint8(i.MsgIntAlgo))

	return b, nil
}
