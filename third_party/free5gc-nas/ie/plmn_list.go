package ie

import (
	"github.com/pkg/errors"
)

type PLMNList struct {
	PlmnIds []PlmnId
}

const (
	MaxNumPlmnId int = 15
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PLMNList) UnmarshalBinary(b []byte) error {
	l := len(b)
	if ((l % PlmnIdPktSz) != 0) || (l < PlmnIdPktSz) || (l > PlmnIdPktSz*MaxNumPlmnId) {
		return errors.Errorf("The PlmnIdDigit IE length(%d) is incorrect", len(b))
	}

	i.PlmnIds = make([]PlmnId, l/PlmnIdPktSz)
	if nil == i.PlmnIds {
		return errors.Errorf("PLMNList:Failed to make %d PlmnIdDigts", l/PlmnIdPktSz)
	}

	for idx := 0; (idx * PlmnIdPktSz) < l; idx++ {
		offset := idx * 3
		if err := i.PlmnIds[idx].UnmarshalBinary(b[offset : offset+3]); err != nil {
			return errors.Wrap(err, "PLMNList UnmarshalBinary")
		}
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PLMNList) MarshalBinary() ([]byte, error) {
	l := len(i.PlmnIds)
	if l > MaxNumPlmnId {
		l = MaxNumPlmnId
	}
	if l == 0 {
		return nil, nil
	}

	b := make([]byte, l*PlmnIdPktSz)
	for idx := 0; idx < l; idx++ {
		offset := idx * 3
		if err := i.PlmnIds[idx].MarshalBinary(b[offset : offset+3]); err != nil {
			return nil, errors.Wrap(err, "PLMNList MarshalBinary")
		}
	}
	return b, nil
}
