package ie

// AdditionalCfgInd is detailed in 9.11.3.74 Additional configuration indication, 24.501
type AdditionalCfgInd struct {
	// Name, uint8, Bits, Octet
	SCMR uint8 // 1->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AdditionalCfgInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "AdditionalCfgInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AdditionalCfgInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "AdditionalCfgInd",
	}
	return nil, e
}
