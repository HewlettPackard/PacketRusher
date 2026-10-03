package ie

import (
	"github.com/pkg/errors"
)

// Capability5GMM is detailed in 9.11.3.1 5GMM capability, 24.501
type Capability5GMM struct {
	Length int // value should be 1-13
	// Name, uint8, Bits, Octet
	SGC          bool // 8 -> 8 , 3 ->  3
	IPHCCPCiot5G bool // 7 -> 7 , 3 ->  3
	N3Data       bool // 6 -> 6 , 3 ->  3
	CPCiot5G     bool // 5 -> 5 , 3 ->  3
	RestrictEC   bool // 4 -> 4 , 3 ->  3
	LPP          bool // 3 -> 3 , 3 ->  3
	HOAttach     bool // 2 -> 2 , 3 ->  3
	S1Mode       bool // 1 -> 1 , 3 ->  3
	RACS         bool // 8 -> 8 , 4 ->  4
	NSSAA        bool // 7 -> 7 , 4 ->  4
	LCS5G        bool // 6 -> 6 , 4 ->  4
	V2XCNPC5     bool // 5 -> 5 , 4 ->  4
	V2XCEPC5     bool // 4 -> 4 , 4 ->  4
	V2X          bool // 3 -> 3 , 4 ->  4
	UPCiot5G     bool // 2 -> 2 , 4 ->  4
	SRVCC5G      bool // 1 -> 1 , 4 ->  4

	ProSeL2Relay5G bool // 8 -> 8 , 5 ->  5
	ProSeDc5G      bool // 7 -> 7 , 5 ->  5
	ProSeDd5G      bool // 6 -> 6 , 5 ->  5
	ER_NSSAI       bool // 5 -> 5 , 5 ->  5
	EHCCPCiot5G    bool // 4 -> 4 , 5 ->  5
	Multipleup     bool // 3 -> 3 , 5 ->  5
	WUSA           bool // 2 -> 2 , 5 ->  5
	CAG            bool // 1 -> 1 , 5 ->  5

	PR             bool // 8 -> 8 , 6 ->  6
	RPR            bool // 7 -> 7 , 6 ->  6
	PIV            bool // 6 -> 6 , 6 ->  6
	NCR            bool // 5 -> 5 , 6 ->  6
	NR_PSSI        bool // 4 -> 4 , 6 ->  6
	ProSeL3rmt5G   bool // 3 -> 3 , 6 ->  6
	ProSeL2rmt5G   bool // 2 -> 2 , 6 ->  6
	ProSeL3relay5G bool // 1 -> 1 , 6 ->  6

	MPSIU      bool // 8 -> 8 , 7 ->  7
	UAS        bool // 7 -> 7 , 7 ->  7
	NSAG       bool // 6 -> 6 , 7 ->  7
	ExCAG      bool // 5 -> 5 , 7 ->  7
	SSNPNSI    bool // 4 -> 4 , 7 ->  7
	EventNotif bool // 3 -> 3 , 7 ->  7
	MINT       bool // 2 -> 2 , 7 ->  7
	NSSRG      bool // 1 -> 1 , 7 ->  7

	RCMAN bool // 2 -> 2 , 8 ->  8
	RCMAP bool // 1 -> 1 , 8 ->  8
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Capability5GMM) UnmarshalBinary(b []byte) error {
	i.Length = len(b)

	// The 5GMM capability is a type 4 information element with a minimum
	// length of 3 octets and a maximum length of 15 octets.
	if i.Length < 1 || i.Length > 13 {
		return errors.Errorf("Bad Capability5GMM  IE length(%d)", i.Length)
	}
	i.SGC = GetBit8(b[0]) == 1
	i.IPHCCPCiot5G = GetBit7(b[0]) == 1
	i.N3Data = GetBit6(b[0]) == 0
	i.CPCiot5G = GetBit5(b[0]) == 1
	i.RestrictEC = GetBit4(b[0]) == 1
	i.LPP = GetBit3(b[0]) == 1
	i.HOAttach = GetBit2(b[0]) == 1
	i.S1Mode = GetBit1(b[0]) == 1
	if i.Length == 1 {
		return nil
	}

	i.RACS = GetBit8(b[1]) == 1
	i.NSSAA = GetBit7(b[1]) == 1
	i.LCS5G = GetBit6(b[1]) == 1
	i.V2XCNPC5 = GetBit5(b[1]) == 1
	i.V2XCEPC5 = GetBit4(b[1]) == 1
	i.V2X = GetBit3(b[1]) == 1
	i.UPCiot5G = GetBit2(b[1]) == 1
	i.SRVCC5G = GetBit1(b[1]) == 1
	if i.Length == 2 {
		return nil
	}

	i.ProSeL2Relay5G = GetBit8(b[2]) == 1
	i.ProSeDc5G = GetBit7(b[2]) == 1
	i.ProSeDd5G = GetBit6(b[2]) == 1
	i.ER_NSSAI = GetBit5(b[2]) == 1
	i.EHCCPCiot5G = GetBit4(b[2]) == 1
	i.Multipleup = GetBit3(b[2]) == 1
	i.WUSA = GetBit2(b[2]) == 1
	i.CAG = GetBit1(b[2]) == 1
	if i.Length == 3 {
		return nil
	}

	i.PR = GetBit8(b[3]) == 1
	i.RPR = GetBit7(b[3]) == 1
	i.PIV = GetBit6(b[3]) == 1
	i.NCR = GetBit5(b[3]) == 1
	i.NR_PSSI = GetBit4(b[3]) == 1
	i.ProSeL3rmt5G = GetBit3(b[3]) == 1
	i.ProSeL2rmt5G = GetBit2(b[3]) == 1
	i.ProSeL3relay5G = GetBit1(b[3]) == 1
	if i.Length == 4 {
		return nil
	}

	i.MPSIU = GetBit8(b[4]) == 1
	i.UAS = GetBit7(b[4]) == 1
	i.NSAG = GetBit6(b[4]) == 1
	i.ExCAG = GetBit5(b[4]) == 1
	i.SSNPNSI = GetBit4(b[4]) == 1
	i.EventNotif = GetBit3(b[4]) == 1
	i.MINT = GetBit2(b[4]) == 1
	i.NSSRG = GetBit1(b[4]) == 1
	if i.Length == 5 {
		return nil
	}

	i.RCMAN = GetBit2(b[5]) == 1
	i.RCMAP = GetBit1(b[5]) == 1

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Capability5GMM) MarshalBinary() ([]byte, error) {
	if i.Length < 1 || i.Length > 13 {
		return nil, errors.Errorf("Bad Capability5GMM  IE length(%d)", i.Length)
	}

	b := make([]byte, i.Length)
	b[0] = SetBit8(b[0], bool2uint8(i.SGC))
	b[0] = SetBit7(b[0], bool2uint8(i.IPHCCPCiot5G))
	b[0] = SetBit6(b[0], bool2uint8(!i.N3Data))
	b[0] = SetBit5(b[0], bool2uint8(i.CPCiot5G))
	b[0] = SetBit4(b[0], bool2uint8(i.RestrictEC))
	b[0] = SetBit3(b[0], bool2uint8(i.LPP))
	b[0] = SetBit2(b[0], bool2uint8(i.HOAttach))
	b[0] = SetBit1(b[0], bool2uint8(i.S1Mode))
	if i.Length == 1 {
		return b, nil
	}

	b[1] = SetBit8(b[1], bool2uint8(i.RACS))
	b[1] = SetBit7(b[1], bool2uint8(i.NSSAA))
	b[1] = SetBit6(b[1], bool2uint8(i.LCS5G))
	b[1] = SetBit5(b[1], bool2uint8(i.V2XCNPC5))
	b[1] = SetBit4(b[1], bool2uint8(i.V2XCEPC5))
	b[1] = SetBit3(b[1], bool2uint8(i.V2X))
	b[1] = SetBit2(b[1], bool2uint8(i.UPCiot5G))
	b[1] = SetBit1(b[1], bool2uint8(i.SRVCC5G))
	if i.Length == 2 {
		return b, nil
	}

	b[2] = SetBit8(b[2], bool2uint8(i.ProSeL2Relay5G))
	b[2] = SetBit7(b[2], bool2uint8(i.ProSeDc5G))
	b[2] = SetBit6(b[2], bool2uint8(i.ProSeDd5G))
	b[2] = SetBit5(b[2], bool2uint8(i.ER_NSSAI))
	b[2] = SetBit4(b[2], bool2uint8(i.EHCCPCiot5G))
	b[2] = SetBit3(b[2], bool2uint8(i.Multipleup))
	b[2] = SetBit2(b[2], bool2uint8(i.WUSA))
	b[2] = SetBit1(b[2], bool2uint8(i.CAG))
	if i.Length == 3 {
		return b, nil
	}

	b[3] = SetBit8(b[3], bool2uint8(i.PR))
	b[3] = SetBit7(b[3], bool2uint8(i.RPR))
	b[3] = SetBit6(b[3], bool2uint8(i.PIV))
	b[3] = SetBit5(b[3], bool2uint8(i.NCR))
	b[3] = SetBit4(b[3], bool2uint8(i.NR_PSSI))
	b[3] = SetBit3(b[3], bool2uint8(i.ProSeL3rmt5G))
	b[3] = SetBit2(b[3], bool2uint8(i.ProSeL2rmt5G))
	b[3] = SetBit1(b[3], bool2uint8(i.ProSeL3relay5G))
	if i.Length == 4 {
		return b, nil
	}

	b[4] = SetBit8(b[4], bool2uint8(i.MPSIU))
	b[4] = SetBit7(b[4], bool2uint8(i.UAS))
	b[4] = SetBit6(b[4], bool2uint8(i.NSAG))
	b[4] = SetBit5(b[4], bool2uint8(i.ExCAG))
	b[4] = SetBit4(b[4], bool2uint8(i.SSNPNSI))
	b[4] = SetBit3(b[4], bool2uint8(i.EventNotif))
	b[4] = SetBit2(b[4], bool2uint8(i.MINT))
	b[4] = SetBit1(b[4], bool2uint8(i.NSSRG))
	if i.Length == 5 {
		return b, nil
	}

	b[5] = SetBit2(b[5], bool2uint8(i.RCMAN))
	b[5] = SetBit1(b[5], bool2uint8(i.RCMAP))
	return b, nil
}
