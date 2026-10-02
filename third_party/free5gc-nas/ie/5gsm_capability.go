package ie

import "github.com/pkg/errors"

// Capability5GSM is detailed in 9.11.4.1 5GSM capability, 24.501
type Capability5GSM struct {
	// Name, uint8, Bits, Octet
	TPMIC   bool  // 8 -> 8 ,   3 -> 3 // True if Transfer of port management information containers supported
	ATSSSST uint8 // 7 -> 4 ,   3 -> 3
	EPTS1   bool  // 3 -> 3 ,   3 -> 3 // True if Ethernet PDN type in S1 mode supported
	MH6PDU  bool  // 2 -> 2 ,   3 -> 3 // True if Multi-homed IPv6 PDU session supported
	Rqos    bool  // 1 -> 1 ,   3 -> 3 // True if Reflective QoS supported
	APMQF   bool  // 1 -> 1 ,   4 -> 4 // true if access performance measurements per QoS flow supported.
}

const (
	AtsssST_AtsssNotSupported                                                  uint8 = 0x00
	AtsssST_ATSSSLowLayerWithAnySteeringModeSupported                          uint8 = 0x01
	AtsssST_MPTCPAnySteeringMode_ATSSSLLOnlyActiveStandbySteeringModeSupported uint8 = 0x02
	AtsssST_MPTCPAnySteeringMode_ATSSSLLAnySteeringModeSupported               uint8 = 0x03
	AtsssST_Reserved1                                                          uint8 = 0x04
	AtsssST_Reserved2                                                          uint8 = 0x05
	AtsssST_Reserved3                                                          uint8 = 0x06
	AtsssST_Reserved4                                                          uint8 = 0x07
	AtsssST_Reserved5                                                          uint8 = 0x08
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Capability5GSM) UnmarshalBinary(b []byte) error {
	if len(b) < 1 {
		return errors.Errorf("The Capability5GSM IE length(%d) is not enough", len(b))
	}
	i.TPMIC = GetBit8(b[0]) == 1
	i.ATSSSST = Get4Bits74(b[0])
	i.EPTS1 = GetBit3(b[0]) == 1
	i.MH6PDU = GetBit2(b[0]) == 1
	i.Rqos = GetBit1(b[0]) == 1
	if len(b) < 2 {
		return nil
	}
	i.APMQF = GetBit1(b[1]) == 1
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Capability5GSM) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2)
	b[0] = SetBit8(b[0], bool2uint8(i.TPMIC))
	b[0] = Set4Bits74(b[0], i.ATSSSST)
	b[0] = SetBit3(b[0], bool2uint8(i.EPTS1))
	b[0] = SetBit2(b[0], bool2uint8(i.MH6PDU))
	b[0] = SetBit1(b[0], bool2uint8(i.Rqos))

	b[1] = SetBit1(b[1], bool2uint8(i.APMQF))

	return b, nil
}
