package ie

// AllowedPDUSessStatus is detailed in 9.11.3.13 Allowed PDU session status, 24.501
type AllowedPDUSessStatus struct {
	Psi
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AllowedPDUSessStatus) UnmarshalBinary(b []byte) error {
	return i.Psi.UnmarshalBinary(b)
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AllowedPDUSessStatus) MarshalBinary() ([]byte, error) {
	return i.Psi.MarshalBinary()
}
