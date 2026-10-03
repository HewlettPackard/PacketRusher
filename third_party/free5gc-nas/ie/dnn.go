package ie

// DNN is detailed in 9.11.2.1B DNN, 24.501
type DNN struct {
	// Name, uint8, Bits, Octet
	Value string // 8->1, 3 -> 100
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DNN) UnmarshalBinary(b []byte) error {
	i.Value = decodeFQDN(b)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DNN) MarshalBinary() ([]byte, error) {
	b := encodeFQDN(i.Value)
	return b, nil
}
