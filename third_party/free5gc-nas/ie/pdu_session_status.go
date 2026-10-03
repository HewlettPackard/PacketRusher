package ie

// PDUSessStatus is detailed in 9.11.3.44 PDU session status, 24.501
type PDUSessStatus struct {
	Psi
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUSessStatus) UnmarshalBinary(b []byte) error {
	return i.Psi.UnmarshalBinary(b)
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUSessStatus) MarshalBinary() ([]byte, error) {
	return i.Psi.MarshalBinary()
}
