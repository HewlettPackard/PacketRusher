package ie

// ECSAddr is detailed in 9.11.4.34 ECS address, 24.501
type ECSAddr struct {
	// Name, uint8, Bits, Octet
	TypeOfECSAddr                    uint8 // 8 -> 4 ,   4 -> 4
	TypeOfSpatialValidityCondition   uint8 // 4 -> 1 ,   4 -> 4
	ECSAddr                          uint8 // 8 -> 1 ,   5 -> 995
	SpatialValidityConditionContents uint8 // 8 -> 1 ,   996 -> 995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ECSAddr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "ECSAddr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ECSAddr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "ECSAddr",
	}
	return nil, e
}
