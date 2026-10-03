package ie

import "github.com/pkg/errors"

// ReqType is detailed in 9.11.3.47 Request type, 24.501
type ReqType struct {
	// Name, uint8, Bits, Octet
	Value ConstReqType // 3 -> 1 ,   1 -> 1
}

type ConstReqType uint8

const (
	ReqType_InitialReq           ConstReqType = 0x01
	ReqType_ExistingPDUSess      ConstReqType = 0x02
	ReqType_InitialEmergReq      ConstReqType = 0x03
	ReqType_ExistingEmergPDUSess ConstReqType = 0x04
	ReqType_ModReq               ConstReqType = 0x05
	ReqType_MAPDUReq             ConstReqType = 0x06
	ReqType_Reserved             ConstReqType = 0x07
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *ReqType) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The ReqType IE length(%d) is incorrect", len(b))
	}
	i.Value = ConstReqType(Get3Bits31(b[0]))
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *ReqType) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits31(b[0], uint8(i.Value))

	return b, nil
}
