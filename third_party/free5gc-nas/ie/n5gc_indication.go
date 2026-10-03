package ie

// N5GCInd is detailed in 9.11.3.72 N5GC indication, 24.501
type N5GCInd struct { // Name, uint8, Bits, Octet
	// special docx fmt in the spec, please hand craft this
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *N5GCInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "N5GCInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *N5GCInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "N5GCInd",
	}
	return nil, e
}
