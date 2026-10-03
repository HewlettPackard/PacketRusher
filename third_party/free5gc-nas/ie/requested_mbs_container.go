package ie

// ReqMBSCntr is detailed in 9.11.4.30 Requested MBS container, 24.501
type ReqMBSCntr struct {
	// Name, uint8, Bits, Octet
	MulticastMBSSessInfo1 uint8 // 8 -> 1 ,   4 -> 995
	MulticastMBSSessInfo2 uint8 // 8 -> 1 ,   996 -> 995
	MulticastMBSSessInfoP uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ReqMBSCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ReqMBSCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ReqMBSCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ReqMBSCntr",
	}

	return nil, e
}
