package ie

// NID is detailed in 9.11.3.79 NID, 24.501
type NID struct {
	// Name, uint8, Bits, Octet
	NIDValueDigit1  uint8 // 8 -> 5 ,   3 -> 3
	AssignmentMode  uint8 // 4 -> 1 ,   3 -> 3
	NIDValueDigit3  uint8 // 8 -> 5 ,   4 -> 4
	NIDValueDigit2  uint8 // 4 -> 1 ,   4 -> 4
	NIDValueDigit5  uint8 // 8 -> 5 ,   5 -> 5
	NIDValueDigit4  uint8 // 4 -> 1 ,   5 -> 5
	NIDValueDigit7  uint8 // 8 -> 5 ,   6 -> 6
	NIDValueDigit6  uint8 // 4 -> 1 ,   6 -> 6
	NIDValueDigit9  uint8 // 8 -> 5 ,   7 -> 7
	NIDValueDigit8  uint8 // 5 -> 5 ,   8 -> 8 (Duplicated?)
	NIDValueDigit10 uint8 // 4 -> 1 ,   8 -> 8
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NID) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "NID",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NID) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "NID",
	}

	return nil, e
}
