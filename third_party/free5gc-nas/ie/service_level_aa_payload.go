package ie

// SvcLvlAAPayload is detailed in 9.11.2.13 Service-level-AA payload, 24.501
type SvcLvlAAPayload struct {
	// Name, uint8, Bits, Octet
	SvcLvlAAPayload uint8 // 8 -> 1 ,   4 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAAPayload) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPayload",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAAPayload) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPayload",
	}

	return nil, e
}
