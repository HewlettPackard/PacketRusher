package ie

// S1ModeToN1ModeNASTransparentCntr is detailed in 9.11.2.9 S1 mode to N1 mode NAS transparent container, 24.501
type S1ModeToN1ModeNASTransparentCntr struct { // Name, uint8, Bits, Octet
	// special docx fmt in the spec, please hand craft this
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *S1ModeToN1ModeNASTransparentCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "S1ModeToN1ModeNASTransparentCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *S1ModeToN1ModeNASTransparentCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "S1ModeToN1ModeNASTransparentCntr",
	}
	return nil, e
}
