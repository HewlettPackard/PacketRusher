package ie

// UERadioCapabilityIDDelInd is detailed in 9.11.3.69 UE radio capability ID deletion indication, 24.501
type UERadioCapabilityIDDelInd struct {
	// Name, uint8, Bits, Octet
	DelReq uint8 // 3->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UERadioCapabilityIDDelInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "UERadioCapabilityIDDelInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UERadioCapabilityIDDelInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "UERadioCapabilityIDDelInd",
	}
	return nil, e
}
