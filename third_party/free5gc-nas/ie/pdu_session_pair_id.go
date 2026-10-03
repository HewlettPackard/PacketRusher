package ie

// PDUSessPairID is detailed in 9.11.4.32 PDU session pair ID, 24.501
type PDUSessPairID struct {
	// Name, uint8, Bits, Octet
	PDUSessPairID uint8 // 8 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUSessPairID) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "PDUSessPairID",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUSessPairID) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "PDUSessPairID",
	}

	return nil, e
}
