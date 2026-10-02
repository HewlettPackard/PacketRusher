package ie

import "github.com/pkg/errors"

// Non3GppNWProvidedPolicies implements the standard single-octet value; spare bits are ignored on decode.
type Non3GppNWProvidedPolicies struct {
	Value uint8
}

func (i *Non3GppNWProvidedPolicies) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("Non3GppNWProvidedPolicies IE length %d, expected 1", len(b))
	}
	i.Value = b[0] & 1
	return nil
}

func (i *Non3GppNWProvidedPolicies) MarshalBinary() ([]byte, error) {
	if i.Value > 1 {
		return nil, errors.Errorf("invalid non-3GPP policies value %d", i.Value)
	}
	return []byte{i.Value}, nil
}
