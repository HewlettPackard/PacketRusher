package ie

// MAPDUSessInfo is detailed in 9.11.3.31A MA PDU session information, 24.501
type MAPDUSessInfo struct {
	// Name, uint8, Bits, Octet
	MAPDUSessInfoValue uint8 // 4->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *MAPDUSessInfo) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "MAPDUSessInfo",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MAPDUSessInfo) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "MAPDUSessInfo",
	}
	return nil, e
}
