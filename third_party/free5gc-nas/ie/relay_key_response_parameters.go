package ie

// RelayKeyRspParams is detailed in 9.11.3.90 Relay key response parameters, 24.501
type RelayKeyRspParams struct {
	// Name, uint8, Bits, Octet
	KeyKnr_Prose uint8 // 8 -> 1 ,   4 -> 35
	Nonce_2      uint8 // 8 -> 1 ,   36 -> 51
	CPPRUKID     uint8 // 8 -> 1 ,   52 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RelayKeyRspParams) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RelayKeyRspParams",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RelayKeyRspParams) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RelayKeyRspParams",
	}

	return nil, e
}
