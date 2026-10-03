package ie

// CiotSmallDataCntr is detailed in 9.11.3.18B CIoT small data container, 24.501
type CiotSmallDataCntr struct {
	// Name, uint8, Bits, Octet
	CiotSmallDataCntrContents bool // 1->1,   4->257
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CiotSmallDataCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "CiotSmallDataCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CiotSmallDataCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "CiotSmallDataCntr",
	}
	return nil, e
}
