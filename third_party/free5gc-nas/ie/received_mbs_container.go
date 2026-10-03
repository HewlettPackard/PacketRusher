package ie

// ReceivedMBSCntr is detailed in 9.11.4.31 Received MBS container, 24.501
type ReceivedMBSCntr struct {
	// Name, uint8, Bits, Octet
	ReceivedMBSInfo1 uint8 // 8 -> 1 ,   4 -> 995
	ReceivedMBSInfo2 uint8 // 8 -> 1 ,   996 -> 995
	ReceivedMBSInfoP uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ReceivedMBSCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ReceivedMBSCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ReceivedMBSCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ReceivedMBSCntr",
	}

	return nil, e
}
