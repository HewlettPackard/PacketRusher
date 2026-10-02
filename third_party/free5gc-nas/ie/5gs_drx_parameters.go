package ie

import (
	"fmt"

	"github.com/pkg/errors"
)

// DRXParams5GS is detailed in 9.11.3.2A 5GS DRX parameters, 24.501
type DRXParams5GS struct {
	// Name, uint8, Bits, Octet
	Value DRXValue // 4->1,   3->3
}

func (i *DRXParams5GS) String() string {
	return fmt.Sprintf("Value: %s", i.Value)
}

type DRXValue uint8

const (
	DRXValueNotSpecified  DRXValue = 0x00
	DRXCycleParameterT32  DRXValue = 0x01
	DRXCycleParameterT64  DRXValue = 0x02
	DRXCycleParameterT128 DRXValue = 0x03
	DRXCycleParameterT256 DRXValue = 0x04
	DRXCycleParameterMAX  DRXValue = DRXCycleParameterT256
)

func (i DRXValue) String() string {
	switch i {
	case DRXValueNotSpecified:
		return "NotSpecified"
	case DRXCycleParameterT32:
		return "T=32"
	case DRXCycleParameterT64:
		return "T=64"
	case DRXCycleParameterT128:
		return "T=128"
	case DRXCycleParameterT256:
		return "T=256"
	default:
		return fmt.Sprintf("Unknown(%d)", i)
	}
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *DRXParams5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("DRXParams5GS buffer len(%d) != 1", len(b))
	}
	if i.Value = DRXValue(Get4Bits41(b[0])); i.Value > DRXCycleParameterMAX {
		i.Value = DRXValueNotSpecified
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *DRXParams5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set4Bits41(b[0], uint8(i.Value))
	return b, nil
}
