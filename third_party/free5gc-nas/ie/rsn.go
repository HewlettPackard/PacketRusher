package ie

// RSN is detailed in 9.11.4.33 RSN, 24.501
type RSN struct {
	// Name, uint8, Bits, Octet
	RSN uint8 // 8 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RSN) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RSN",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RSN) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RSN",
	}

	return nil, e
}
