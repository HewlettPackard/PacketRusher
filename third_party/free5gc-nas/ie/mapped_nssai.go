package ie

// MappedNSSAI is detailed in 9.11.3.31B Mapped NSSAI, 24.501
type MappedNSSAI struct {
	// Name, uint8, Bits, Octet
	MappedSNSSAIContent1 uint8 // 8->1,   3->995
	MappedSNSSAIContent2 uint8 // 8->1,   996->1
	MappedSNSSAIContentN uint8 // 8->1,   2->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *MappedNSSAI) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "MappedNSSAI",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MappedNSSAI) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "MappedNSSAI",
	}
	return nil, e
}
