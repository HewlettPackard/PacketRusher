package ie

import (
	"github.com/pkg/errors"
)

// UpdateType5GS is detailed in 9.11.3.9A 5GS update type, 24.501
type UpdateType5GS struct {
	// Name, uint8, Bits, Octet
	EPS_PNB_CIoT uint8 // 6 -> 5 ,   1 -> 1 EPS Preferred CIoT network behaviour
	PNB_CIoT_5GS uint8 // 4 -> 3 ,   1 -> 1 5GS Preferred CIoT network behaviour
	NG_RAN_RCU   bool  // 2 -> 2 ,   1 -> 1 True if UE radio capability update needed
	SMSRequested bool  // 1 -> 1 ,   1 -> 1 True if SMS over NAS supported
}

// TS 24.501 9.11.3.9A
const (
	// EPS_PNB_CIoT , PNB_CIoT_5GS
	PNB_CIoT_NoAdditionalInfo             uint8 = 0
	PNB_CIoT_CtrlPlaneCiot5GSOptimization uint8 = 1
	PNB_CIoT_UserPlaneCiot5GSOptimization uint8 = 2

	// NG_RAN_RCU
	UERadioCapabilityUpdateNotNeeded bool = false
	UERadioCapabilityUpdateNeeded    bool = true

	// SMSRequested
	SMSOverNASNotSupported bool = false
	SMSOverNASSupported    bool = true
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *UpdateType5GS) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The UpdateType5GS  IE length(%d) is incorrect", len(b))
	}
	i.EPS_PNB_CIoT = Get2Bits65(b[0])
	i.PNB_CIoT_5GS = Get2Bits43(b[0])
	i.NG_RAN_RCU = GetBit2(b[0]) == 1
	i.SMSRequested = GetBit1(b[0]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *UpdateType5GS) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set2Bits65(b[0], i.EPS_PNB_CIoT)
	b[0] = Set2Bits43(b[0], i.PNB_CIoT_5GS)
	b[0] = SetBit2(b[0], bool2uint8(i.NG_RAN_RCU))
	b[0] = SetBit1(b[0], bool2uint8(i.SMSRequested))

	return b, nil
}
