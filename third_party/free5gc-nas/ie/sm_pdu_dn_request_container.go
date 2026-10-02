package ie

// SMPDUDNReqCntr is detailed in 9.11.4.15 SM PDU DN request container, 24.501
type SMPDUDNReqCntr struct {
	// Name, uint8, Bits, Octet
	DNSpecificId uint8 // 8->1,   3->3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SMPDUDNReqCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SMPDUDNReqCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SMPDUDNReqCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SMPDUDNReqCntr",
	}
	return nil, e
}
