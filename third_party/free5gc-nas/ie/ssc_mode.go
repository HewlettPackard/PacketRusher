package ie

import "github.com/pkg/errors"

// SSCMode is detailed in 9.11.4.16 SSC Mode, 24.501
type SSCMode struct {
	Mode uint8
}

const (
	SSCMODE1 uint8 = iota + 1
	SSCMODE2
	SSCMODE3
)

func (i *SSCMode) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("SSCMode bad IE len(%d)", len(b))
	}
	i.Mode = Get3Bits31(b[0])
	return nil
}

func (i *SSCMode) MarshalBinary() ([]byte, error) {
	return []byte{i.Mode}, nil
}
