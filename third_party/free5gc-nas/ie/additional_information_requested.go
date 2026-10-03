package ie

// AdditionalInfoReq is detailed in 9.11.3.12A Additional information requested, 24.501
type AdditionalInfoReq struct {
	// Name, uint8, Bits, Octet
	Cipherkey uint8 // 1->1,   3->3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *AdditionalInfoReq) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "AdditionalInfoReq",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *AdditionalInfoReq) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "AdditionalInfoReq",
	}
	return nil, e
}
