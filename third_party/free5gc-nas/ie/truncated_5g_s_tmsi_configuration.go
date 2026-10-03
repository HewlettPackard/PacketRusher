package ie

// Truncated5GSTMSICfg is detailed in 9.11.3.70 Truncated 5G-S-TMSI configuration, 24.501
type Truncated5GSTMSICfg struct {
	// Name, uint8, Bits, Octet
	TruncatedAMFSetIDValue   uint8 // 8->5,   3->3
	TruncatedAMFPointerValue uint8 // 5->1,   3->3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Truncated5GSTMSICfg) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "Truncated5GSTMSICfg",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Truncated5GSTMSICfg) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "Truncated5GSTMSICfg",
	}
	return nil, e
}
