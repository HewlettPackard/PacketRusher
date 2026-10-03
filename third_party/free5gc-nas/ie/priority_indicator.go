package ie

// PriorityIndicator is detailed in 9.11.3.91 Priority indicator, 24.501
type PriorityIndicator struct {
	// Name, uint8, Bits, Octet
	MPSI uint8 // 1 -> 1 ,   1 -> 1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PriorityIndicator) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "PriorityIndicator",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PriorityIndicator) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "PriorityIndicator",
	}

	return nil, e
}
