package ie

// UEParamsUpdateTransparentCntr is detailed in 9.11.3.53A UE parameters update transparent container, 24.501
type UEParamsUpdateTransparentCntr struct {
	// Name, uint8, Bits, Octet
	UEParamsUpdateHdr  uint8 // 8->1,   4->4
	UPUMACIAUSF        uint8 // 8->1,   5->520
	Counterupu         uint8 // 8->1,   521->2122
	UEParamsUpdateList uint8 // 8->1,   2123->23
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UEParamsUpdateTransparentCntr) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "UEParamsUpdateTransparentCntr",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UEParamsUpdateTransparentCntr) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "UEParamsUpdateTransparentCntr",
	}
	return nil, e
}
