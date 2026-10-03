package ie

// CAGInfoList is detailed in 9.11.3.18A CAG information list, 24.501
type CAGInfoList struct {
	// Name, uint8, Bits, Octet
	Entry1 uint8 // 8->1,   4->995
	Entry2 uint8 // 8->1,   996->1
	EntryN uint8 // 8->1,   2->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CAGInfoList) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "CAGInfoList",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CAGInfoList) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "CAGInfoList",
	}
	return nil, e
}
