package ie

// PDUSessReactivationResult is detailed in 9.11.3.42 PDU session reactivation result, 24.501
type PDUSessReactivationResult struct {
	Psi
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUSessReactivationResult) UnmarshalBinary(b []byte) error {
	return i.Psi.UnmarshalBinary(b)
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUSessReactivationResult) MarshalBinary() ([]byte, error) {
	return i.Psi.MarshalBinary()
}
