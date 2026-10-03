package ie

// SvcLvlAAPendingInd is detailed in 9.11.2.17 Service-level-AA pending indication, 24.501
type SvcLvlAAPendingInd struct {
	// Name, uint8, Bits, Octet
	SLAPI uint8 // 1 -> 1 ,   1 -> 1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAAPendingInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPendingInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAAPendingInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPendingInd",
	}

	return nil, e
}
