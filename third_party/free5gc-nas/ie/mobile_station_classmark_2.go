package ie

type MobileStationClassmark2 struct{}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *MobileStationClassmark2) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "MobileStationClassmark2",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *MobileStationClassmark2) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "MobileStationClassmark2",
	}
	return nil, e
}
