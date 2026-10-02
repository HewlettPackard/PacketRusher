package ie

// UplinkDataStatus is detailed in 9.11.3.57 Uplink data status, 24.501
type UplinkDataStatus struct {
	Psi
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UplinkDataStatus) UnmarshalBinary(b []byte) error {
	return i.Psi.UnmarshalBinary(b)
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UplinkDataStatus) MarshalBinary() ([]byte, error) {
	return i.Psi.MarshalBinary()
}
