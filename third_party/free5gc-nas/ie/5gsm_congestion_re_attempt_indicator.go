package ie

import "github.com/pkg/errors"

// CongestionReattemptIndicator5GSM implements the standard single-octet value; spare bits are ignored on decode.
type CongestionReattemptIndicator5GSM struct {
	ABO   uint8
	CATBO uint8
}

func (i *CongestionReattemptIndicator5GSM) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("CongestionReattemptIndicator5GSM IE length %d, expected 1", len(b))
	}
	i.ABO = GetBit1(b[0])
	i.CATBO = GetBit2(b[0])
	return nil
}

func (i *CongestionReattemptIndicator5GSM) MarshalBinary() ([]byte, error) {
	if i.ABO > 1 || i.CATBO > 1 {
		return nil, errors.Errorf("invalid congestion re-attempt bits")
	}
	return []byte{i.CATBO<<1 | i.ABO}, nil
}
