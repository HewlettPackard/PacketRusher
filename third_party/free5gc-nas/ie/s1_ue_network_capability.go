package ie

import (
	"github.com/pkg/errors"
)

// S1UENwCapability is detailed in 9.11.3.48 S1 UE network capability, 24.501
// which references to  9.9.3.34 UE network capability, 24.301
type S1UENwCapability struct {
	// EPS encryption algorithms supported
	EEA0     bool // 8 -> 8 , 3 -> 3
	EEA1_128 bool // 7 -> 7 , 3 -> 3
	EEA2_128 bool // 6 -> 6 , 3 -> 3
	EEA3_128 bool // 5 -> 5 , 3 -> 3
	EEA4     bool // 4 -> 4 , 3 -> 3
	EEA5     bool // 3 -> 3 , 3 -> 3
	EEA6     bool // 2 -> 2 , 3 -> 3
	EEA7     bool // 1 -> 1 , 3 -> 3

	// EPS integrity algorithms supported
	EIA0     bool // 8 -> 8 , 4 -> 4
	EIA1_128 bool // 7 -> 7 , 4 -> 4
	EIA2_128 bool // 6 -> 6 , 4 -> 4
	EIA3_128 bool // 5 -> 5 , 4 -> 4
	EIA4     bool // 4 -> 4 , 4 -> 4
	EIA5     bool // 3 -> 3 , 4 -> 4
	EIA6     bool // 2 -> 2 , 4 -> 4
	EIA7     bool // 1 -> 1 , 4 -> 4

	// UMTS encryption algorithms supported
	UEA0 bool // 8 -> 8 , 5 -> 5
	UEA1 bool // 7 -> 7 , 5 -> 5
	UEA2 bool // 6 -> 6 , 5 -> 5
	UEA3 bool // 5 -> 5 , 5 -> 5
	UEA4 bool // 4 -> 4 , 5 -> 5
	UEA5 bool // 3 -> 3 , 5 -> 5
	UEA6 bool // 2 -> 2 , 5 -> 5
	UEA7 bool // 1 -> 1 , 5 -> 5

	// UMTS integrity algorithms supported
	UCS2 bool // 8 -> 8 , 6 -> 6
	UIA1 bool // 7 -> 7 , 6 -> 6
	UIA2 bool // 6 -> 6 , 6 -> 6
	UIA3 bool // 5 -> 5 , 6 -> 6
	UIA4 bool // 4 -> 4 , 6 -> 6
	UIA5 bool // 3 -> 3 , 6 -> 6
	UIA6 bool // 2 -> 2 , 6 -> 6
	UIA7 bool // 1 -> 1 , 6 -> 6

	ProSe_dd bool // 8 -> 8 , 7 -> 7
	ProSe    bool // 7 -> 7 , 7 -> 7
	H245_ASH bool // 6 -> 6 , 7 -> 7
	ACC_CSFB bool // 5 -> 5 , 7 -> 7
	LPP      bool // 4 -> 4 , 7 -> 7
	LCS      bool // 3 -> 3 , 7 -> 7
	SRVCC_1x bool // 2 -> 2 , 7 -> 7
	NF       bool // 1 -> 1 , 7 -> 7

	EPCO       bool // 8 -> 8 , 8 -> 8
	HC_CP_CIoT bool // 7 -> 7 , 8 -> 8
	ERw_oPDN   bool // 6 -> 6 , 8 -> 8
	S1U_Data   bool // 5 -> 5 , 8 -> 8
	UP_CIoT    bool // 4 -> 4 , 8 -> 8
	CP_CIoT    bool // 3 -> 3 , 8 -> 8
	ProseRelay bool // 2 -> 2 , 8 -> 8
	ProSe_dc   bool // 1 -> 1 , 8 -> 8

	Bearers15   bool // 8 -> 8 , 9 -> 9
	SGC         bool // 7 -> 7 , 9 -> 9
	N1mode      bool // 6 -> 6 , 9 -> 9
	DCNR        bool // 5 -> 5 , 9 -> 9
	CP_Backoff  bool // 4 -> 4 , 9 -> 9
	RestrictEC  bool // 3 -> 3 , 9 -> 9
	V2X_PC5     bool // 2 -> 2 , 9 -> 9
	MultipleDRB bool // 1 -> 1 , 9 -> 9

	V2X_NR_PC5 bool // 5 -> 5 , 10 -> 10
	UP_MT_EDT  bool // 4 -> 4 , 10 -> 10
	CP_MT_EDT  bool // 3 -> 3 , 10 -> 10
	WUSA       bool // 2 -> 2 , 10 -> 10
	RACS       bool // 1 -> 1 , 10 -> 10
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *S1UENwCapability) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return errors.Errorf("Bad S1UENwCapability IE length(%d)", len(b))
	} else {
		i.EEA0 = GetBit8(b[0]) == 1
		i.EEA1_128 = GetBit7(b[0]) == 1
		i.EEA2_128 = GetBit6(b[0]) == 1
		i.EEA3_128 = GetBit5(b[0]) == 1
		i.EEA4 = GetBit4(b[0]) == 1
		i.EEA5 = GetBit3(b[0]) == 1
		i.EEA6 = GetBit2(b[0]) == 1
		i.EEA7 = GetBit1(b[0]) == 1
		i.EIA0 = GetBit8(b[1]) == 1
		i.EIA1_128 = GetBit7(b[1]) == 1
		i.EIA2_128 = GetBit6(b[1]) == 1
		i.EIA3_128 = GetBit5(b[1]) == 1
		i.EIA4 = GetBit4(b[1]) == 1
		i.EIA5 = GetBit3(b[1]) == 1
		i.EIA6 = GetBit2(b[1]) == 1
		i.EIA7 = GetBit1(b[1]) == 1
	}
	if len(b) > 2 {
		i.UEA0 = GetBit8(b[2]) == 1
		i.UEA1 = GetBit7(b[2]) == 1
		i.UEA2 = GetBit6(b[2]) == 1
		i.UEA3 = GetBit5(b[2]) == 1
		i.UEA4 = GetBit4(b[2]) == 1
		i.UEA5 = GetBit3(b[2]) == 1
		i.UEA6 = GetBit2(b[2]) == 1
		i.UEA7 = GetBit1(b[2]) == 1
	}
	if len(b) > 3 {
		i.UCS2 = GetBit8(b[3]) == 1
		i.UIA1 = GetBit7(b[3]) == 1
		i.UIA2 = GetBit6(b[3]) == 1
		i.UIA3 = GetBit5(b[3]) == 1
		i.UIA4 = GetBit4(b[3]) == 1
		i.UIA5 = GetBit3(b[3]) == 1
		i.UIA6 = GetBit2(b[3]) == 1
		i.UIA7 = GetBit1(b[3]) == 1
	}
	if len(b) > 4 {
		i.ProSe_dd = GetBit8(b[4]) == 1
		i.ProSe = GetBit7(b[4]) == 1
		i.H245_ASH = GetBit6(b[4]) == 1
		i.ACC_CSFB = GetBit5(b[4]) == 1
		i.LPP = GetBit4(b[4]) == 1
		i.LCS = GetBit3(b[4]) == 1
		i.SRVCC_1x = GetBit2(b[4]) == 1
		i.NF = GetBit1(b[4]) == 1
	}
	if len(b) > 5 {
		i.EPCO = GetBit8(b[5]) == 1
		i.HC_CP_CIoT = GetBit7(b[5]) == 1
		i.ERw_oPDN = GetBit6(b[5]) == 1
		i.S1U_Data = GetBit5(b[5]) == 1
		i.UP_CIoT = GetBit4(b[5]) == 1
		i.CP_CIoT = GetBit3(b[5]) == 1
		i.ProseRelay = GetBit2(b[5]) == 1
		i.ProSe_dc = GetBit1(b[5]) == 1
	}
	if len(b) > 6 {
		i.Bearers15 = GetBit8(b[6]) == 1
		i.SGC = GetBit7(b[6]) == 1
		i.N1mode = GetBit6(b[6]) == 1
		i.DCNR = GetBit5(b[6]) == 1
		i.CP_Backoff = GetBit4(b[6]) == 1
		i.RestrictEC = GetBit3(b[6]) == 1
		i.V2X_PC5 = GetBit2(b[6]) == 1
		i.MultipleDRB = GetBit1(b[6]) == 1
	}
	if len(b) > 7 {
		i.V2X_NR_PC5 = GetBit5(b[7]) == 1
		i.UP_MT_EDT = GetBit4(b[7]) == 1
		i.CP_MT_EDT = GetBit3(b[7]) == 1
		i.WUSA = GetBit2(b[7]) == 1
		i.RACS = GetBit1(b[7]) == 1
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *S1UENwCapability) MarshalBinary() ([]byte, error) {
	b := make([]byte, 8)
	b[0] = SetBit8(b[0], bool2uint8(i.EEA0))
	b[0] = SetBit7(b[0], bool2uint8(i.EEA1_128))
	b[0] = SetBit6(b[0], bool2uint8(i.EEA2_128))
	b[0] = SetBit5(b[0], bool2uint8(i.EEA3_128))
	b[0] = SetBit4(b[0], bool2uint8(i.EEA4))
	b[0] = SetBit3(b[0], bool2uint8(i.EEA5))
	b[0] = SetBit2(b[0], bool2uint8(i.EEA6))
	b[0] = SetBit1(b[0], bool2uint8(i.EEA7))
	b[1] = SetBit8(b[1], bool2uint8(i.EIA0))
	b[1] = SetBit7(b[1], bool2uint8(i.EIA1_128))
	b[1] = SetBit6(b[1], bool2uint8(i.EIA2_128))
	b[1] = SetBit5(b[1], bool2uint8(i.EIA3_128))
	b[1] = SetBit4(b[1], bool2uint8(i.EIA4))
	b[1] = SetBit3(b[1], bool2uint8(i.EIA5))
	b[1] = SetBit2(b[1], bool2uint8(i.EIA6))
	b[1] = SetBit1(b[1], bool2uint8(i.EIA7))
	b[2] = SetBit8(b[2], bool2uint8(i.UEA0))
	b[2] = SetBit7(b[2], bool2uint8(i.UEA1))
	b[2] = SetBit6(b[2], bool2uint8(i.UEA2))
	b[2] = SetBit5(b[2], bool2uint8(i.UEA3))
	b[2] = SetBit4(b[2], bool2uint8(i.UEA4))
	b[2] = SetBit3(b[2], bool2uint8(i.UEA5))
	b[2] = SetBit2(b[2], bool2uint8(i.UEA6))
	b[2] = SetBit1(b[2], bool2uint8(i.UEA7))
	b[3] = SetBit8(b[3], bool2uint8(i.UCS2))
	b[3] = SetBit7(b[3], bool2uint8(i.UIA1))
	b[3] = SetBit6(b[3], bool2uint8(i.UIA2))
	b[3] = SetBit5(b[3], bool2uint8(i.UIA3))
	b[3] = SetBit4(b[3], bool2uint8(i.UIA4))
	b[3] = SetBit3(b[3], bool2uint8(i.UIA5))
	b[3] = SetBit2(b[3], bool2uint8(i.UIA6))
	b[3] = SetBit1(b[3], bool2uint8(i.UIA7))
	b[4] = SetBit8(b[4], bool2uint8(i.ProSe_dd))
	b[4] = SetBit7(b[4], bool2uint8(i.ProSe))
	b[4] = SetBit6(b[4], bool2uint8(i.H245_ASH))
	b[4] = SetBit5(b[4], bool2uint8(i.ACC_CSFB))
	b[4] = SetBit4(b[4], bool2uint8(i.LPP))
	b[4] = SetBit3(b[4], bool2uint8(i.LCS))
	b[4] = SetBit2(b[4], bool2uint8(i.SRVCC_1x))
	b[4] = SetBit1(b[4], bool2uint8(i.NF))
	b[5] = SetBit8(b[5], bool2uint8(i.EPCO))
	b[5] = SetBit7(b[5], bool2uint8(i.HC_CP_CIoT))
	b[5] = SetBit6(b[5], bool2uint8(i.ERw_oPDN))
	b[5] = SetBit5(b[5], bool2uint8(i.S1U_Data))
	b[5] = SetBit4(b[5], bool2uint8(i.UP_CIoT))
	b[5] = SetBit3(b[5], bool2uint8(i.CP_CIoT))
	b[5] = SetBit2(b[5], bool2uint8(i.ProseRelay))
	b[5] = SetBit1(b[5], bool2uint8(i.ProSe_dc))
	b[6] = SetBit8(b[6], bool2uint8(i.Bearers15))
	b[6] = SetBit7(b[6], bool2uint8(i.SGC))
	b[6] = SetBit6(b[6], bool2uint8(i.N1mode))
	b[6] = SetBit5(b[6], bool2uint8(i.DCNR))
	b[6] = SetBit4(b[6], bool2uint8(i.CP_Backoff))
	b[6] = SetBit3(b[6], bool2uint8(i.RestrictEC))
	b[6] = SetBit2(b[6], bool2uint8(i.V2X_PC5))
	b[6] = SetBit1(b[6], bool2uint8(i.MultipleDRB))
	b[7] = SetBit5(b[7], bool2uint8(i.V2X_NR_PC5))
	b[7] = SetBit4(b[7], bool2uint8(i.UP_MT_EDT))
	b[7] = SetBit3(b[7], bool2uint8(i.CP_MT_EDT))
	b[7] = SetBit2(b[7], bool2uint8(i.WUSA))
	b[7] = SetBit1(b[7], bool2uint8(i.RACS))

	return b, nil
}
