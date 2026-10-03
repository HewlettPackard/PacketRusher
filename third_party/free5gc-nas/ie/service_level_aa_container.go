package ie

// SvcLvlAACntr is detailed in 9.11.2.10 Service-level-AA container, 24.501
type SvcLvlAACntr struct {
	// Name, uint8, Bits, Octet
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAACntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SvcLvlAACntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAACntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SvcLvlAACntr",
	}

	return nil, e
}
