package ie

// ExtendedRejectedNSSAI is detailed in 9.11.3.75 Extended rejected NSSAI, 24.501
type ExtendedRejectedNSSAI struct {
	// Name, uint8, Bits, Octet
	PartialExtendedRejectedNSSAIList1 uint8 // 8 -> 1 ,   3 -> 995
	PartialExtendedRejectedNSSAIList2 uint8 // 8 -> 1 ,   996 -> 995
	PartialExtendedRejectedNSSAIListN uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ExtendedRejectedNSSAI) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ExtendedRejectedNSSAI",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ExtendedRejectedNSSAI) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ExtendedRejectedNSSAI",
	}

	return nil, e
}
