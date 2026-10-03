package ie

// PagingRestriction is detailed in 9.11.3.77 Paging restriction, 24.501
type PagingRestriction struct {
	// Name, uint8, Bits, Octet
	PagingRestrictionType uint8 // 4 -> 1 ,   3 -> 3
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PagingRestriction) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "PagingRestriction",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PagingRestriction) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "PagingRestriction",
	}

	return nil, e
}

type PagingRestriction_1 struct {
	// Name, uint8, Bits, Octet
	PagingRestrictionType uint8 // 4 -> 1 ,   3 -> 3
	PSI7                  uint8 // 8 -> 8 ,   4 -> 4
	PSI6                  uint8 // 7 -> 7 ,   4 -> 4
	PSI5                  uint8 // 6 -> 6 ,   4 -> 4
	PSI4                  uint8 // 5 -> 5 ,   4 -> 4
	PSI3                  uint8 // 4 -> 4 ,   4 -> 4
	PSI2                  uint8 // 3 -> 3 ,   4 -> 4
	PSI1                  uint8 // 2 -> 2 ,   4 -> 4
	PSI0                  uint8 // 1 -> 1 ,   4 -> 4
	PSI15                 uint8 // 8 -> 8 ,   5 -> 5
	PSI14                 uint8 // 7 -> 7 ,   5 -> 5
	PSI13                 uint8 // 6 -> 6 ,   5 -> 5
	PSI12                 uint8 // 5 -> 5 ,   5 -> 5
	PSI11                 uint8 // 4 -> 4 ,   5 -> 5
	PSI10                 uint8 // 3 -> 3 ,   5 -> 5
	PSI9                  uint8 // 2 -> 2 ,   5 -> 5
	PSI8                  uint8 // 1 -> 1 ,   6 -> 995 (Duplicated?)
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PagingRestriction_1) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "PagingRestriction",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PagingRestriction_1) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "PagingRestriction",
	}

	return nil, e
}
