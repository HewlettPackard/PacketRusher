package ie

// IPHdrCompressionCfg is detailed in 9.11.4.24 IP header compression configuration, 24.501
type IPHdrCompressionCfg struct {
	// Name, uint8, Bits, Octet
	P0X0104                                      bool  // 7->7,   3->3
	P0X0103                                      bool  // 6->6,   3->3
	P0X0102                                      bool  // 5->5,   3->3
	P0X0006                                      bool  // 4->4,   3->3
	P0X0004                                      bool  // 3->3,   3->3
	P0X0003                                      bool  // 2->2,   3->3
	P0X0002                                      bool  // 1->1,   3->3
	MAX_CID                                      uint8 // 8->1,   4->5
	AdditionalIPHdrCompressionCtxSetupParamsType uint8 // 8->1,   6->6
	AdditionalIPHdrCompressionCtxSetupParamsCont uint8 // 8->1,   7->995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *IPHdrCompressionCfg) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "IPHdrCompressionCfg",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *IPHdrCompressionCfg) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "IPHdrCompressionCfg",
	}
	return nil, e
}
