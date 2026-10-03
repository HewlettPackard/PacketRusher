package ie

// CipheringKeyData is detailed in 9.11.3.18C Ciphering key data, 24.501
type CipheringKeyData struct {
	// Name, uint8, Bits, Octet
	CipheringDataSet1 uint8 // 8->1,   4->995
	CipheringDataSet2 uint8 // 8->1,   996->1
	CipheringDataSetP uint8 // 8->1,   2->1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CipheringKeyData) UnmarshalBinary(b []byte) error {
	var e error = &IEToDo{
		IEName: "CipheringKeyData",
	}
	return e
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CipheringKeyData) MarshalBinary() ([]byte, error) {
	var e error = &IEToDo{
		IEName: "CipheringKeyData",
	}
	return nil, e
}
