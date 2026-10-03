package ie

import (
	"github.com/pkg/errors"
)

// NASMsgCntr is detailed in 9.11.3.33 NAS message container, 24.501
type NASMsgCntr struct {
	// Name, uint8, Bits, Octet
	Contents []byte // 1 -> 1 ,   4 -> n
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NASMsgCntr) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("The NASMsgCntr IE length (%d) is incorrect", len(b))
	}
	i.Contents = make([]byte, 0, len(b))
	if nil == i.Contents {
		return errors.Errorf("NASMsgCntr: failed to make %d []byte", len(b))
	}
	i.Contents = append(i.Contents, b...)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NASMsgCntr) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, len(i.Contents))
	b = append(b, i.Contents...)

	return b, nil
}
