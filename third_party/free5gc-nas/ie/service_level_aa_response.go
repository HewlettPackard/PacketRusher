package ie

// SvcLvlAARsp is detailed in 9.11.2.14 Service-level-AA response, 24.501
type SvcLvlAARsp struct {
	// Name, uint8, Bits, Octet
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAARsp) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SvcLvlAARsp",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAARsp) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SvcLvlAARsp",
	}

	return nil, e
}
