package ie

type RelAssistanceInd struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RelAssistanceInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RelAssistanceInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RelAssistanceInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RelAssistanceInd",
	}
	return nil, e
}
