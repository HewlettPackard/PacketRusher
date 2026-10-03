package ie

// ProseRelayTransactionId is detailed in 9.11.3.88 ProSe relay transaction identity, 24.501
type ProseRelayTransactionId struct {
	// Name, uint8, Bits, Octet
	ProseRelayTransactionIdValue uint8 // 8 -> 1 ,   2 -> 2
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ProseRelayTransactionId) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ProseRelayTransactionId",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ProseRelayTransactionId) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ProseRelayTransactionId",
	}

	return nil, e
}
