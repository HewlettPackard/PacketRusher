package ie

// EPSNASMsgCntr is detailed in 9.11.3.24 EPS NAS message container, 24.501
type EPSNASMsgCntr struct {
	// Name, uint8, Bits, Octet
	EPSNASMsgCntr uint8 // 8->1,   4->995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *EPSNASMsgCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "EPSNASMsgCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *EPSNASMsgCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "EPSNASMsgCntr",
	}
	return nil, e
}
