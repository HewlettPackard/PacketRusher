package ie

// ATSSSCntr is detailed in 9.11.4.22 ATSSS container, 24.501
type ATSSSCntr struct {
	// Name, uint8, Bits, Octet
	ATSSSCntrContents uint8 // 8->1,   4->995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ATSSSCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ATSSSCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ATSSSCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ATSSSCntr",
	}
	return nil, e
}
