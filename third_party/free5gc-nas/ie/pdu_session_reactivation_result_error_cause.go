package ie

// PDUSessReactivationResultErrCause is detailed in 9.11.3.43 PDU session reactivation result error cause, 24.501
type PDUSessReactivationResultErrCause struct {
	IdCause []SessIdCausePair
}

type SessIdCausePair struct {
	// Name, uint8, Bits, Octet
	PDUSessID uint8 // 8->1,   4->4
	Value     uint8 // 8->1,   5->5
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PDUSessReactivationResultErrCause) UnmarshalBinary(b []byte) error {
	sz := len(b)
	ofs := 0
	for ofs+1 < sz {
		pair := SessIdCausePair{PDUSessID: b[ofs], Value: b[ofs+1]}
		i.IdCause = append(i.IdCause, pair)
		ofs += 2
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PDUSessReactivationResultErrCause) MarshalBinary() ([]byte, error) {
	sz := len(i.IdCause) * 2
	b := make([]byte, sz)

	ofs := 0
	for _, pair := range i.IdCause {
		b[ofs] = pair.PDUSessID
		b[ofs+1] = pair.Value
		ofs += 2
	}
	return b, nil
}
