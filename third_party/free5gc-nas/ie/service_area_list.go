package ie

import "github.com/pkg/errors"

// SvcAreaList is detailed in 9.11.3.49 Service area list, 24.501
// This IE is very similar to 9.11.3.9 5GS tracking area identity list, 24.501
type SvcAreaList struct {
	AllAllowed  bool
	AllowedList TrackingAreaIdList5GS
	DeniedList  TrackingAreaIdList5GS
}

const (
	// Type of list. For type 0~2, Refer to ie/5gs_tracking_area_identity_list.go
	ConstAllTAIsInAllowedArea uint8 = 3

	AllowedArea    uint8 = 0
	NonAllowedArea uint8 = 1
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcAreaList) UnmarshalBinary(b []byte) error {
	if len(b) < 4 {
		return errors.Errorf("SvcAreaList: bad len (%d) to unmarshal", len(b))
	}
	deny := GetBit8(b[0]) == 1
	listType := Get2Bits76(b[0])
	if listType == ConstAllTAIsInAllowedArea {
		// how to handle this ?
		i.AllAllowed = true
		return nil
	}
	if deny {
		if err := i.DeniedList.UnmarshalBinary(b); err != nil {
			return errors.Wrap(err, "SvcAreaList Denied UnmarshalBinary()")
		}
	} else {
		if err := i.AllowedList.UnmarshalBinary(b); err != nil {
			return errors.Wrap(err, "SvcAreaList Allowed UnmarshalBinary()")
		}
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcAreaList) MarshalBinary() ([]byte, error) {
	var allowed, denied []byte
	var err error

	// TS 24.501, Table 9.11.3.4.1 'The "Allowed type" fields in all the
	// partial service area lists shall have the same value.'
	// means no mixed allow/non-allowed lists?
	if len(i.AllowedList.TAI) > 0 {
		allowed, err = i.AllowedList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAreaList Allowed MarshalBinary()")
		}
		allowed[0] = SetBit8(allowed[0], AllowedArea)
	} else if len(i.DeniedList.TAI) > 0 {
		denied, err = i.DeniedList.MarshalBinary()
		if err != nil {
			return nil, errors.Wrap(err, "SvcAreaList Denied MarshalBinary()")
		}
		denied[0] = SetBit8(denied[0], NonAllowedArea)
	}

	return append(allowed, denied...), nil
}
