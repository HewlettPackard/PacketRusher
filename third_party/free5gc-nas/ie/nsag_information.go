package ie

// NSAGInfo is detailed in 9.11.3.87 NSAG information, 24.501
type NSAGInfo struct {
	// Name, uint8, Bits, Octet
	NSAG1 uint8 // 8 -> 1 ,   4 -> 995
	NSAG2 uint8 // 8 -> 1 ,   996 -> 995
	NSAGX uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NSAGInfo) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "NSAGInfo",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NSAGInfo) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "NSAGInfo",
	}

	return nil, e
}
