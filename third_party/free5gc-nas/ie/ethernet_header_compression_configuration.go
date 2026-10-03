package ie

type EthHdrCompressionCfg struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *EthHdrCompressionCfg) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "EthHdrCompressionCfg",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *EthHdrCompressionCfg) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "EthHdrCompressionCfg",
	}
	return nil, e
}
