package ie

import "github.com/pkg/errors"

// CtrlPlaneSvcType is detailed in 9.11.3.18D Control plane service type, 24.501
type CtrlPlaneSvcType struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 3 -> 1 ,   1 -> 1
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *CtrlPlaneSvcType) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("the CtrlPlaneSvcType IE len(%d) < 1", len(b))
	}
	i.Value = Get3Bits31(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *CtrlPlaneSvcType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits31(b[0], i.Value)

	return b, nil
}
