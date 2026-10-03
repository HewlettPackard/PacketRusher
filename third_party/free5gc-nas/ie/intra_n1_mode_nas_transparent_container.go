package ie

// IntraN1ModeNASTransparentCntr is detailed in 9.11.2.6 Intra N1 mode NAS transparent container, 24.501
type IntraN1ModeNASTransparentCntr struct {
	// Name, uint8, Bits, Octet
	MsgAuthCode                   uint8 // 8->1,   3->6
	TypeOfCipheringAlgo           uint8 // 8->5,   7->7
	TypeOfIntegrityProtectionAlgo uint8 // 4->1,   7->7
	KACF                          bool  // 5->5,   8->8
	TSC                           bool  // 4->4,   8->8
	KeySetIdentifierIn5G          uint8 // 3->1,   8->8
	SequenceNum                   uint8 // 8->1,   9->9
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *IntraN1ModeNASTransparentCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "IntraN1ModeNASTransparentCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *IntraN1ModeNASTransparentCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "IntraN1ModeNASTransparentCntr",
	}
	return nil, e
}
