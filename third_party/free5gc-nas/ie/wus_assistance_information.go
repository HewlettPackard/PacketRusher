package ie

type WUSAssistanceInfo struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *WUSAssistanceInfo) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "WUSAssistanceInfo",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *WUSAssistanceInfo) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "WUSAssistanceInfo",
	}
	return nil, e
}
