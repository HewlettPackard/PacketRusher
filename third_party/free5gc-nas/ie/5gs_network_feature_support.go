package ie

import "github.com/pkg/errors"

// NwFeatureSupport5GS is detailed in 9.11.3.5 5GS network feature support, 24.501
type NwFeatureSupport5GS struct {
	Length uint8

	// Name, uint8, Bits, Octet
	MPSI         bool  // 8 -> 8 ,   3 -> 3   MPS indicator, whether access identity 1 is valid.
	IWKN26       bool  // 7 -> 7 ,   3 -> 3   Interwork w/o N26 interface support
	EMF          uint8 // 6 -> 5 ,   3 -> 3
	EMC          uint8 // 4 -> 3 ,   3 -> 3
	IMSVoPSN3GPP bool  // 2 -> 2 ,   3 -> 3   True if IMS voice over PS session supported over Non3GPP access
	IMSVoPS3GPP  bool  // 1 -> 1 ,   3 -> 3   True if IMS voice over PS session supported over 3GPP access
	UPCiot5G     bool  // 8 -> 8 ,   4 -> 4   True if User plane CIoT 5GS optimization supported
	IPHCCPCiot5G bool  // 7 -> 7 ,   4 -> 4   True if IP hdr compression for ctrl plane CIoT 5GS optimization is supported
	N3Data       bool  // 6 -> 6 ,   4 -> 4   True if N3 data transfer is *NOT* supported
	CPCiot5G     bool  // 5 -> 5 ,   4 -> 4   True if Control plane CIoT 5GS optimization is supported
	RestrictEC   uint8 // 4 -> 3 ,   4 -> 4   Restriction on enhanced coverage
	MCSI         bool  // 2 -> 2 ,   4 -> 4   MCS indicator, whether Access identity 2 is valid
	EMCN3        bool  // 1 -> 1 ,   4 -> 4   True if Emergency service support for non-3GPP access
	PR           bool  // 7 -> 7 ,   5 -> 5   True if paging restriction supported
	RPR          bool  // 6 -> 6 ,   5 -> 5   True if reject paging request supported
	PIV          bool  // 5 -> 5 ,   5 -> 5   True if paging indication for voice services supported
	NCR          bool  // 4 -> 4 ,   5 -> 5   True if N1-NAS signalling connection release supported
	EHCCPCiot5G  bool  // 3 -> 3 ,   5 -> 5   True if Eth hdr compression for Ctrl plane CIoT 5GS optimization supported
	ATSIND       bool  // 2 -> 2 ,   5 -> 5   True if ATSSS is supported
	LCS5G        bool  // 1 -> 1 ,   5 -> 5   True if Location Services in 5GC supported
}

const (
	// EMC
	EmergSvc_NotSupported                               uint8 = 0
	EmergSvc_InNRConnectedTo5GCNOnly                    uint8 = 1
	EmergSvc_InEUTRAConnectedTo5GCNOnly                 uint8 = 2
	EmergSvc_InNRConnectedTo5GCNAndEUTRAConnectedTo5GCN uint8 = 3

	// EMF
	EmergSvcFallback_NotSupported                               uint8 = 0
	EmergSvcFallback_InNRConnectedTo5GCNOnly                    uint8 = 1
	EmergSvcFallback_InEUTRAConnectedTo5GCNOnly                 uint8 = 2
	EmergSvcFallback_InNRConnectedTo5GCNAndEUTRAConnectedTo5GCN uint8 = 3

	// RestrictEC / WB-N1 ; for NB-N1, 1 means Restricted.
	Restricted_None            uint8 = 0
	Restricted_CEModeA_CEModeB uint8 = 1
	Restricted_CEModeB         uint8 = 2
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *NwFeatureSupport5GS) UnmarshalBinary(b []byte) error {
	if len(b) < 1 || len(b) > 3 {
		return errors.Errorf("The NwFeatureSupport5GS IE length(%d) is incorrect", len(b))
	}

	i.Length = uint8(len(b))
	i.MPSI = GetBit8(b[0]) == 1
	i.IWKN26 = GetBit7(b[0]) == 1
	i.EMF = Get2Bits65(b[0])
	i.EMC = Get2Bits43(b[0])
	i.IMSVoPSN3GPP = GetBit2(b[0]) == 1
	i.IMSVoPS3GPP = GetBit1(b[0]) == 1
	if i.Length > 1 {
		i.UPCiot5G = GetBit8(b[1]) == 1
		i.IPHCCPCiot5G = GetBit7(b[1]) == 1
		i.N3Data = GetBit6(b[1]) == 1
		i.CPCiot5G = GetBit5(b[1]) == 1
		i.RestrictEC = Get2Bits43(b[1])
		i.MCSI = GetBit2(b[1]) == 1
		i.EMCN3 = GetBit1(b[1]) == 1
	}
	if i.Length > 2 {
		i.PR = GetBit7(b[2]) == 1
		i.RPR = GetBit6(b[2]) == 1
		i.PIV = GetBit5(b[2]) == 1
		i.NCR = GetBit4(b[2]) == 1
		i.EHCCPCiot5G = GetBit3(b[2]) == 1
		i.ATSIND = GetBit2(b[2]) == 1
		i.LCS5G = GetBit1(b[2]) == 1
	}

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *NwFeatureSupport5GS) MarshalBinary() ([]byte, error) {
	if i.Length < 1 || i.Length > 3 {
		return nil, errors.Errorf("The NwFeatureSupport5GS IE length(%d) is incorrect", i.Length)
	}

	b := make([]byte, i.Length)
	b[0] = SetBit8(b[0], bool2uint8(i.MPSI))
	b[0] = SetBit7(b[0], bool2uint8(i.IWKN26))
	b[0] = Set2Bits65(b[0], i.EMF)
	b[0] = Set2Bits43(b[0], i.EMC)
	b[0] = SetBit2(b[0], bool2uint8(i.IMSVoPSN3GPP))
	b[0] = SetBit1(b[0], bool2uint8(i.IMSVoPS3GPP))
	if i.Length > 1 {
		b[1] = SetBit8(b[1], bool2uint8(i.UPCiot5G))
		b[1] = SetBit7(b[1], bool2uint8(i.IPHCCPCiot5G))
		b[1] = SetBit6(b[1], bool2uint8(i.N3Data))
		b[1] = SetBit5(b[1], bool2uint8(i.CPCiot5G))
		b[1] = Set2Bits43(b[1], i.RestrictEC)
		b[1] = SetBit2(b[1], bool2uint8(i.MCSI))
		b[1] = SetBit1(b[1], bool2uint8(i.EMCN3))
	}
	if i.Length > 2 {
		b[2] = SetBit7(b[2], bool2uint8(i.PR))
		b[2] = SetBit6(b[2], bool2uint8(i.RPR))
		b[2] = SetBit5(b[2], bool2uint8(i.PIV))
		b[2] = SetBit4(b[2], bool2uint8(i.NCR))
		b[2] = SetBit3(b[2], bool2uint8(i.EHCCPCiot5G))
		b[2] = SetBit2(b[2], bool2uint8(i.ATSIND))
		b[2] = SetBit1(b[2], bool2uint8(i.LCS5G))
	}

	return b, nil
}
