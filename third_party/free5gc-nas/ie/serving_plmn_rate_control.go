package ie

type ServingPLMNRateCtrl struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ServingPLMNRateCtrl) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ServingPLMNRateCtrl",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ServingPLMNRateCtrl) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ServingPLMNRateCtrl",
	}
	return nil, e
}
