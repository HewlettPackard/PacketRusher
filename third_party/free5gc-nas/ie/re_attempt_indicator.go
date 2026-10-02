package ie

// ReattemptIndicator is detailed in 9.11.4.17 Re-attempt indicator, 24.501
type ReattemptIndicator struct {
	// Name, uint8, Bits, Octet
	EPLMNC bool // 2->1,   3->3
	RATC   bool // 1->1,   3->3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ReattemptIndicator) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ReattemptIndicator",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ReattemptIndicator) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ReattemptIndicator",
	}
	return nil, e
}
