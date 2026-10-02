package ie

// DisasterPlmnList is detailed in 9.11.3.83 List of PLMNs to be used in disaster condition, 24.501
type DisasterPlmnList struct {
	// Name, uint8, Bits, Octet
	PLMNID1 uint8 // 8 -> 1 ,   3 -> 5
	PLMNID2 uint8 // 8 -> 1 ,   6 -> 8
	// PLMNID2 uint8 // 1 -> 1 ,   9 -> 995 (Duplicated?)
	PLMNIDN uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DisasterPlmnList) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "DisasterPlmnList",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DisasterPlmnList) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "DisasterPlmnList",
	}
	return nil, e
}
