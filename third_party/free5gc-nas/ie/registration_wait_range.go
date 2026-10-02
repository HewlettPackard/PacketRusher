package ie

// RegWaitRange is detailed in 9.11.3.84 Registration wait range, 24.501
type RegWaitRange struct {
	// Name, uint8, Bits, Octet
	MinRegWaitTime uint8 // 8 -> 1 ,   3 -> 3
	MaxRegWaitTime uint8 // 8 -> 1 ,   4 -> 4
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RegWaitRange) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RegWaitRange",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RegWaitRange) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RegWaitRange",
	}

	return nil, e
}
