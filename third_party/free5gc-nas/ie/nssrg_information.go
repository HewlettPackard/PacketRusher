package ie

// NSSRGInfo is detailed in 9.11.3.82 NSSRG information, 24.501
type NSSRGInfo struct {
	// Name, uint8, Bits, Octet
	NSSRGValuesForSNSSAI1 uint8 // 8 -> 1 ,   4 -> 995
	NSSRGValuesForSNSSAI2 uint8 // 8 -> 1 ,   996 -> 995
	NSSRGValuesForSNSSAIX uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NSSRGInfo) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "NSSRGInfo",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NSSRGInfo) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "NSSRGInfo",
	}

	return nil, e
}
