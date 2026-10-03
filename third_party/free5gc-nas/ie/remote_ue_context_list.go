package ie

// RemoteUECtxList is detailed in 9.11.4.29 Remote UE context list, 24.501
type RemoteUECtxList struct {
	// Name, uint8, Bits, Octet
	NumOfRemoteUECtxs uint8 // 8 -> 1 ,   4 -> 4
	RemoteUECtx1      uint8 // 8 -> 1 ,   5 -> 5
	RemoteUECtxK      uint8 // 8 -> 1 ,   996 -> 995 (Duplicated?)
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *RemoteUECtxList) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "RemoteUECtxList",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *RemoteUECtxList) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "RemoteUECtxList",
	}

	return nil, e
}
