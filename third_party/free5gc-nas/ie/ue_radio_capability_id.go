package ie

// UERadioCapabilityID is detailed in 9.11.3.68 UE radio capability ID, 24.501
type UERadioCapabilityID struct {
	// Name, uint8, Bits, Octet
	UERadioCapabilityID uint8 // 8->1,   3->995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UERadioCapabilityID) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "UERadioCapabilityID",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UERadioCapabilityID) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "UERadioCapabilityID",
	}
	return nil, e
}
