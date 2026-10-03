package ie

import "github.com/pkg/errors"

// MaxNumOfSupportedPktFilters is detailed in 9.11.4.9 Maximum number of supported packet filters, 24.501
type MaxNumOfSupportedPktFilters struct {
	// Name, uint8, Bits, Octet
	MaxNumOfSupportedPktFilters uint8 // 8 -> 6 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *MaxNumOfSupportedPktFilters) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("the MaxNumOfSupportedPktFilters IE len(%d) < 1", len(b))
	}
	i.MaxNumOfSupportedPktFilters = Get3Bits86(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MaxNumOfSupportedPktFilters) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits86(b[0], i.MaxNumOfSupportedPktFilters)

	return b, nil
}
