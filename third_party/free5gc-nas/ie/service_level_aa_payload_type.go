package ie

// SvcLvlAAPayloadType is detailed in 9.11.2.15 Service-level-AA payload type, 24.501
type SvcLvlAAPayloadType struct {
	// Name, uint8, Bits, Octet
	SvcLvlAAPayloadType uint8 // 8 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *SvcLvlAAPayloadType) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPayloadType",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *SvcLvlAAPayloadType) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "SvcLvlAAPayloadType",
	}

	return nil, e
}
