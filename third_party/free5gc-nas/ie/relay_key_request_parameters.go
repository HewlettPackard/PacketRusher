package ie

// RelayKeyReqParams is detailed in 9.11.3.89 Relay key request parameters, 24.501
type RelayKeyReqParams struct {
	// Name, uint8, Bits, Octet
	RelaySvcCode uint8 // 8 -> 1 ,   4 -> 6
	Nonce_1      uint8 // 8 -> 1 ,   7 -> 22
	RemoteUEId   uint8 // 8 -> 1 ,   23 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RelayKeyReqParams) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RelayKeyReqParams",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RelayKeyReqParams) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RelayKeyReqParams",
	}

	return nil, e
}
