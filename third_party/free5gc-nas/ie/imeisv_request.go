package ie

import "github.com/pkg/errors"

// IMEISVReq is detailed in 10.5.5.10 IMEISV request, 24.008
type IMEISVReq struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 3 -> 1 ,   1 -> 1
}

const (
	IMEISV_NotRequested uint8 = 0x00
	IMEISV_Requested    uint8 = 0x01
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *IMEISVReq) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The IMEISVReq  IE length(%d) is incorrect", len(b))
	}
	i.Value = Get3Bits31(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *IMEISVReq) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits31(b[0], i.Value)

	return b, nil
}
