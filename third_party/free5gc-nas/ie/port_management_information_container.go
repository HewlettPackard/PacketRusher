package ie

// PortMgmtInfoCntr is detailed in 9.11.4.27 Port management information container, 24.501
type PortMgmtInfoCntr struct {
	// Name, uint8, Bits, Octet
	PmicCntr []byte // 8->1,   4->995
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PortMgmtInfoCntr) UnmarshalBinary(b []byte) error {
	i.PmicCntr = append(i.PmicCntr, b...)

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PortMgmtInfoCntr) MarshalBinary() ([]byte, error) {
	dataLen := len(i.PmicCntr)
	b := make([]byte, dataLen)
	copy(b, i.PmicCntr)

	return b, nil
}
