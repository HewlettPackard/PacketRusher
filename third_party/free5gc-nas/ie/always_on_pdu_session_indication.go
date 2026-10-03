package ie

import "github.com/pkg/errors"

// AlwaysonPDUSessInd implements the standard single-octet value; spare bits are ignored on decode.
type AlwaysonPDUSessInd struct {
	APSI bool
}

func (i *AlwaysonPDUSessInd) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("AlwaysonPDUSessInd IE length %d, expected 1", len(b))
	}
	i.APSI = GetBit1(b[0]) == 1
	return nil
}

func (i *AlwaysonPDUSessInd) MarshalBinary() ([]byte, error) {

	return []byte{bool2uint8(i.APSI)}, nil
}
