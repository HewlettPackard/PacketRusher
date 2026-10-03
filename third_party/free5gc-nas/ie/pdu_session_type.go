package ie

import "github.com/pkg/errors"

// PDUSessType is detailed in 9.11.4.11 PDU Session Type, 24.501
type PDUSessType struct {
	Value uint8
}

const (
	PDUSessType_IPv4         uint8 = 1
	PDUSessType_IPv6         uint8 = 2
	PDUSessType_IPv4v6       uint8 = 3
	PDUSessType_Unstructured uint8 = 4
	PDUSessType_Ethernet     uint8 = 5
	PDUSessType_Reserved     uint8 = 7
)

var PDUSessTypeStrings = map[uint8]string{
	PDUSessType_IPv4:         "IPv4",
	PDUSessType_IPv6:         "IPv6",
	PDUSessType_IPv4v6:       "IPv4v6",
	PDUSessType_Unstructured: "Unstructured",
	PDUSessType_Ethernet:     "Ethernet",
	PDUSessType_Reserved:     "Reserved",
}

func (i *PDUSessType) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("the PDUSessType IE length(%d) is not enough", len(b))
	}
	i.Value = Get3Bits31(b[0])
	return nil
}

func (i *PDUSessType) MarshalBinary() ([]byte, error) {
	return []byte{i.Value}, nil
}

func (i *PDUSessType) String() string {
	if s, ok := PDUSessTypeStrings[i.Value]; ok {
		return s
	}
	return "Unknown"
}
