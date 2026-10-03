package ie

import "github.com/pkg/errors"

// MICOInd implements the standard single-octet value; spare bits are ignored on decode.
type MICOInd struct {
	SPRTI bool
	RAAI  bool
}

func (i *MICOInd) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("MICOInd IE length %d, expected 1", len(b))
	}
	i.SPRTI = GetBit2(b[0]) == 1
	i.RAAI = GetBit1(b[0]) == 1
	return nil
}

func (i *MICOInd) MarshalBinary() ([]byte, error) {

	return []byte{SetBit2(SetBit1(0, bool2uint8(i.RAAI)), bool2uint8(i.SPRTI))}, nil
}
