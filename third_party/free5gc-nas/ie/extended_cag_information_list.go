package ie

// ExtendedCAGInfoList is detailed in 9.11.3.86 Extended CAG information list, 24.501
type ExtendedCAGInfoList struct {
	// Name, uint8, Bits, Octet
	Entry1 uint8 // 8 -> 1 ,   4 -> 995
	Entry2 uint8 // 8 -> 1 ,   996 -> 995
	EntryN uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ExtendedCAGInfoList) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ExtendedCAGInfoList",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ExtendedCAGInfoList) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ExtendedCAGInfoList",
	}

	return nil, e
}
