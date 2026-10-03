package ie

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// PLMNId is detailed in 9.11.3.85 PLMN identity, 24.501
// This is a common structure in 24.501, 24.008
type PlmnId struct {
	MCC string // Hex string of 3 bytes
	MNC string // Hex string of 3 bytes
}

const (
	PlmnIdPktSz int = 3
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PlmnId) UnmarshalBinary(b []byte) error {
	if len(b) < 3 {
		return errors.Errorf("PlmnId IE len(%d) < 3", len(b))
	}
	var mcc [3]byte
	var mnc [3]byte
	mcc[1], mcc[0] = GetHalfValue(b[0])
	mnc[2], mcc[2] = GetHalfValue(b[1])
	mnc[1], mnc[0] = GetHalfValue(b[2])
	i.MCC = fmt.Sprintf("%d%d%d", mcc[0], mcc[1], mcc[2])
	if mnc[2] == 0xf {
		i.MNC = fmt.Sprintf("%d%d", mnc[0], mnc[1])
	} else {
		i.MNC = fmt.Sprintf("%d%d%d", mnc[0], mnc[1], mnc[2])
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PlmnId) MarshalBinary(b []byte) error {
	if len(b) != 3 {
		return errors.Errorf("PlmnId Tac len(%d) != 3", len(b))
	}

	// MCC
	var mcc [3]byte
	if tmp, err := strconv.Atoi(i.MCC); err != nil {
		return errors.Wrap(err, "PlmnId MCC MarshalBinary")
	} else if len(i.MNC) < 2 || len(i.MNC) > 3 {
		return errors.Errorf("PlmnId MarshalBinary : mcc (%s) range error", i.MCC)
	} else {
		mcc[2] = uint8(tmp % 10)
		tmp /= 10
		mcc[1] = uint8(tmp % 10)
		tmp /= 10
		mcc[0] = uint8(tmp % 10)
	}

	// MNC
	var mnc [3]byte
	if tmp, err := strconv.Atoi(i.MNC); err != nil {
		return errors.Wrap(err, "PlmnId MNC MarshalBinary")
	} else if len(i.MNC) < 2 || len(i.MNC) > 3 {
		return errors.Errorf("PlmnId MarshalBinary : mnc (%s) range error", i.MNC)
	} else {
		if len(i.MNC) == 3 {
			mnc[2] = uint8(tmp % 10)
			tmp /= 10
		} else {
			mnc[2] = uint8(0xf)
		}
		mnc[1] = uint8(tmp % 10)
		tmp /= 10
		mnc[0] = uint8(tmp % 10)
	}

	// compose up the buffer
	b[0] = SetHalfValue(mcc[1], mcc[0])
	b[1] = SetHalfValue(mnc[2], mcc[2])
	b[2] = SetHalfValue(mnc[1], mnc[0])

	return nil
}
