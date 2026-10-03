package ie

// N1ModeToS1ModeNASTransparentCntr is detailed in 9.11.2.7 N1 mode to S1 mode NAS transparent container, 24.501
type N1ModeToS1ModeNASTransparentCntr struct {
	// Name, uint8, Bits, Octet
	SequenceNum uint8 // 8->1,   2->2
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *N1ModeToS1ModeNASTransparentCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "N1ModeToS1ModeNASTransparentCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *N1ModeToS1ModeNASTransparentCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "N1ModeToS1ModeNASTransparentCntr",
	}
	return nil, e
}
