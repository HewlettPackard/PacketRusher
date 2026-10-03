package ie

// PEIPSAssistanceInfo is detailed in 9.11.3.80 PEIPS assistance information, 24.501
type PEIPSAssistanceInfo struct {
	// Name, uint8, Bits, Octet
	PEIPSAssistanceInfoType1 uint8 // 8 -> 1 ,   3 -> 995
	PEIPSAssistanceInfoType2 uint8 // 8 -> 1 ,   996 -> 995
	PEIPSAssistanceInfoTypeP uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PEIPSAssistanceInfo) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "PEIPSAssistanceInfo",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PEIPSAssistanceInfo) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "PEIPSAssistanceInfo",
	}

	return nil, e
}
