package ie

type SupportedCodecList struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SupportedCodecList) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SupportedCodecList",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SupportedCodecList) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SupportedCodecList",
	}
	return nil, e
}
