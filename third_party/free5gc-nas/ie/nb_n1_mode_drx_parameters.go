package ie

// NBN1ModeDRXParams is detailed in 9.11.3.73 NB-N1 mode DRX parameters, 24.501
type NBN1ModeDRXParams struct {
	// Name, uint8, Bits, Octet
	NBN1ModeDRXValue uint8 // 4->1,   3->3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NBN1ModeDRXParams) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "NBN1ModeDRXParams",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NBN1ModeDRXParams) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "NBN1ModeDRXParams",
	}
	return nil, e
}
