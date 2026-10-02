package ie

// CtrlPlaneOnlyInd is detailed in 9.11.4.23 Control plane only indication, 24.501
type CtrlPlaneOnlyInd struct {
	// Name, uint8, Bits, Octet
	CPOIValue bool // 1->1,   1->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CtrlPlaneOnlyInd) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "CtrlPlaneOnlyInd",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CtrlPlaneOnlyInd) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "CtrlPlaneOnlyInd",
	}
	return nil, e
}
