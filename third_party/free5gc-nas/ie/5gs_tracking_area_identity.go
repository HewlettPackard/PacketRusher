package ie

import (
	"encoding/hex"

	"github.com/pkg/errors"
)

// TrackingAreaId5GS is detailed in 9.11.3.8 5GS tracking area identity, 24.501
type TrackingAreaId5GS struct {
	PlmnId
	TAC string // Hex string of 3 bytes, e.g. 001122
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *TrackingAreaId5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 6 {
		return errors.Errorf("The TrackingAreaId5GS IE length(%d) is incorrect", len(b))
	}
	if err := i.UnmarshalPlmnId(b[0:3]); err != nil {
		return errors.Wrap(err, "UnmarshalBinary plmnId")
	}
	if err := i.UnmarshalTac(b[3:6]); err != nil {
		return errors.Wrap(err, "UnmarshalBinary tac")
	}
	return nil
}

func (i *TrackingAreaId5GS) UnmarshalPlmnId(b []byte) error {
	if len(b) != 3 {
		return errors.Errorf("TrackingAreaId5GS PlmnId len(%d) != 3", len(b))
	}
	if err := i.PlmnId.UnmarshalBinary(b); err != nil {
		return errors.Wrap(err, "UnmarshalPlmnId")
	}
	return nil
}

func (i *TrackingAreaId5GS) UnmarshalTac(b []byte) error {
	if len(b) != 3 {
		return errors.Errorf("TrackingAreaId5GS Tac len(%d) != 3", len(b))
	}
	i.TAC = hex.EncodeToString(b[:3])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *TrackingAreaId5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 6)

	if err := i.MarshalPlmnId(b[:3]); err != nil {
		return nil, errors.Wrap(err, "MarshalBinary plmnId")
	}
	if err := i.MarshalTac(b[3:6]); err != nil {
		return nil, errors.Wrap(err, "MarshalBinary tac")
	}
	return b, nil
}

func (i *TrackingAreaId5GS) MarshalPlmnId(b []byte) error {
	if len(b) != 3 {
		return errors.Errorf("TrackingAreaId5GS Tac len(%d) != 3", len(b))
	}
	if err := i.PlmnId.MarshalBinary(b); err != nil {
		return errors.Wrap(err, "MarshalPlmnId")
	}

	return nil
}

func (i *TrackingAreaId5GS) MarshalTac(b []byte) error {
	if len(b) != 3 {
		return errors.Errorf("TrackingAreaId5GS Tac len(%d) != 3", len(b))
	}
	if tac, err := hex.DecodeString(i.TAC); err != nil {
		return errors.Wrap(err, "MarshalTac")
	} else {
		copy(b[0:3], tac)
	}
	return nil
}
