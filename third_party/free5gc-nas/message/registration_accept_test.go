package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestRegAcceptUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1 - 6 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.MobileId5GS                       0x77 TLV-E,      14B, 9.11.3.4
				0x77, 0x00, 0x0b,
				0xf2, 0x02, 0xf8, 0x39, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01,
				// ie.TrackingAreaIdList5GS                 0x54   TLV,   9-114B, 9.11.3.9
				0x54, 0x07,
				0x00, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x01,
				// ie.NSSAI                             0x15   TLV,    4-74B, 9.11.3.37
				0x15, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.NwFeatureSupport5GS               0x21   TLV,     3-5B, 9.11.3.5
				0x21, 0x02, 0x00, 0x00,
				// ie.GPRSTimer3                        0x5E   TLV,       3B, 9.11.2.5
				0x5e, 0x01, 0x06,
				// ie.GPRSTimer2                        0x16   TLV,       3B, 9.11.2.4
				0x16, 0x01, 0x2c,
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				GUTI5G: &ie.MobileId5GS{
					OddEvenIndic: 0,
					TypeOfId:     ie.IdType_5GS_GUTI,
					AllOneBits:   0x0F,
					PlmnId: ie.PlmnId{
						MCC: "208",
						MNC: "93",
					},
					AMFRegionID: 202, AMFSetID: 1016, AMFPointer: 0,
					AMFId:  "cafe00",
					TMSI5G: [4]byte{0x00, 0x00, 0x00, 0x01},
				},
				TAIList: &ie.TrackingAreaIdList5GS{
					TAI: []ie.TrackingAreaId5GS{{
						PlmnId: ie.PlmnId{
							MCC: "208",
							MNC: "93",
						},
						TAC: "000001",
					}},
				},
				AllowedNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				NwFeatureSupport5GS: &ie.NwFeatureSupport5GS{
					Length: 2,
					MPSI:   false, IWKN26: false, EMF: 0x00, EMC: 0x00,
					IMSVoPSN3GPP: false, IMSVoPS3GPP: false, UPCiot5G: false, IPHCCPCiot5G: false,
					N3Data: false, CPCiot5G: false, RestrictEC: 0x00, MCSI: false,
					EMCN3: false, EHCCPCiot5G: false, ATSIND: false, LCS5G: false,
				},
				T3512Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				T3502Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
			},
		},
		{
			name: "Case 2 - 5 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.PLMNList                          0x4A   TLV,    5-47B, 9.11.3.45
				0x4A, 0x06, 0x76, 0x54, 0x32, 0x98, 0x76, 0x54,
				// ie.RejectedNSSAI                     0x11   TLV,    4-42B, 9.11.3.46
				// ie.NSSAI                             0x31   TLV,   4-146B, 9.11.3.37
				0x31, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.PDUSessStatus                     0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0xaa,
				// ie.PDUSessReactivationResult         0x26   TLV,    4-34B, 9.11.3.42
				0x26, 0x02, 0xAA, 0x55,
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				EquivalentPlmns: &ie.PLMNList{
					PlmnIds: []ie.PlmnId{{
						MCC: "674",
						MNC: "235",
					}, {
						MCC: "896",
						MNC: "457",
					}},
				},
				// RejectedNSSAI   : &ie.RejectedNSSAI{},
				ConfiguredNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, true, false, true, false, true, false, true,
					}},
				},
				PDUSessReactivationResult: &ie.PDUSessReactivationResult{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, true, false, true, false, true,
						true, false, true, false, true, false, true, false,
					}},
				},
			},
		},
		{
			name: "Case 3 - 5 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.PDUSessReactivationResultErrCause 0x72 TLV-E,   5-515B, 9.11.3.43
				// ie.LADNInfo                          0x79 TLV-E, 12-1715B, 9.11.3.30
				// ie.MICOInd                           0xB0    TV,       1B, 9.11.3.31
				// ie.NwSlicingInd                      0x90    TV,       1B, 9.11.3.36
				0x92,
				// ie.SvcAreaList                       0x27   TLV,   6-114B, 9.11.3.49
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// PDUSessReactivationResultErrCause : &ie.PDUSessReactivationResultErrCause{},
				// LADNInfo                          : &ie.LADNInfo{},
				// MICOInd                           : &ie.MICOInd{},
				NwSlicingInd: &ie.NwSlicingInd{
					DCNI: true, NSSCI: false,
				},
				// SvcAreaList                       : &ie.SvcAreaList{},
			},
		},
		{
			name: "Case 4 - 5 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.GPRSTimer2                        0x5D   TLV,       3B, 9.11.2.4
				0x5D, 0x01, 0x2c,
				// ie.EmergNumList                        0x34   TLV,    5-50B, 9.11.3.23
				// ie.ExtendedEmergNumList                0x7A TLV-E, 7-65538B, 9.11.3.26
				// ie.SORTransparentCntr                         0x73 TLV-E,    20-nB, 9.11.3.51
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				Non3GppDeregTimerValue: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				// EmergNumList                       : &ie.EmergNumList{},
				// ExtendedEmergNumList               : &ie.ExtendedEmergNumList{},
				// SORTransparentCntr                        : &ie.SORTransparentCntr{},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
			},
		},
		{
			name: "Case 5 - 5 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.ExtendedDRXParams                 0x6E   TLV,       3B, 9.11.3.26A
				// ie.GPRSTimer3                        0x6C   TLV,       3B, 9.11.2.5
				0x6C, 0x01, 0x06,
				// ie.GPRSTimer2                        0x6B   TLV,       3B, 9.11.2.4
				0x6B, 0x01, 0x56,
				// Rejected NSSAI
				// ie.GPRSTimer3                        0x6A   TLV,       3B, 9.11.2.5
				0x6A, 0x01, 0x06,
				// ie.UERadioCapabilityID               0x67   TLV,     3-nB, 9.11.3.68
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// NegotiatedExtendedDRXParams      : &ie.ExtendedDRXParams{},
				T3447Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				T3448Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_Decihours,
					Value: 0x16,
				},
				T3324Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				// UERadioCapabilityID              : &ie.UERadioCapabilityID{},
			},
		},
		{
			name: "Case 6 - 5 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.UERadioCapabilityIDDelInd         0xE0    TV,       1B, 9.11.3.69
				// ie.NSSAI                             0x39   TLV,   4-146B, 9.11.3.37
				0x39, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.CipheringKeyData                  0x74 TLV-E,    34-nB, 9.11.3.18C
				// ie.CAGInfoList                       0x75 TLV-E,     3-nB, 9.11.3.18A
				// ie.Truncated5GSTMSICfg               0x1B   TLV,       3B, 9.11.3.70
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// UERadioCapabilityIDDelInd        : &ie.UERadioCapabilityIDDelInd{},
				PendingNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				// CipheringKeyData                 : &ie.CipheringKeyData{},
				// CAGInfoList                      : &ie.CAGInfoList{},
				// Truncated5GSTMSICfg              : &ie.Truncated5GSTMSICfg{},
			},
		},
		{
			name: "Case 7 - 2 Optional IEs",
			input: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.WUSAssistanceInfo                 0x1C   TLV,     3-nB, 9.11.3.71
				// ie.NBN1ModeDRXParams                 0x29   TLV,       3B, 9.11.3.73
			},
			expected: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// NegotiatedWUSAssistanceInfo      : &ie.WUSAssistanceInfo{},
				// NegotiatedNBN1ModeDRXParams      : &ie.NBN1ModeDRXParams{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(RegAccept)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestRegAcceptMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1 - 6 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.MobileId5GS                       0x77 TLV-E,      14B, 9.11.3.4
				0x77, 0x00, 0x0b,
				0xf2, 0x02, 0xf8, 0x39, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01,
				// ie.TrackingAreaIdList5GS                 0x54   TLV,   9-114B, 9.11.3.9
				0x54, 0x07,
				0x00, 0x02, 0xf8, 0x39, 0x00, 0x00, 0x01,
				// ie.NSSAI                             0x15   TLV,    4-74B, 9.11.3.37
				0x15, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.NwFeatureSupport5GS               0x21   TLV,     3-5B, 9.11.3.5
				0x21, 0x03, 0x00, 0x00, 0x00,
				// ie.GPRSTimer3                        0x5E   TLV,       3B, 9.11.2.5
				0x5e, 0x01, 0x06,
				// ie.GPRSTimer2                        0x16   TLV,       3B, 9.11.2.4
				0x16, 0x01, 0x2c,
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				GUTI5G: &ie.MobileId5GS{
					OddEvenIndic: 0,
					TypeOfId:     ie.IdType_5GS_GUTI,
					AllOneBits:   0x0F,
					PlmnId: ie.PlmnId{
						MCC: "208",
						MNC: "93",
					},
					AMFRegionID: 202, AMFSetID: 1016, AMFPointer: 0,
					TMSI5G: [4]byte{0x00, 0x00, 0x00, 0x01},
				},
				TAIList: &ie.TrackingAreaIdList5GS{
					TAI: []ie.TrackingAreaId5GS{{
						PlmnId: ie.PlmnId{MCC: "208", MNC: "93"},
						TAC:    "000001",
					}},
				},
				AllowedNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				NwFeatureSupport5GS: &ie.NwFeatureSupport5GS{
					Length: 3,
					MPSI:   false, IWKN26: false, EMF: 0x00, EMC: 0x00,
					IMSVoPSN3GPP: false, IMSVoPS3GPP: false, UPCiot5G: false, IPHCCPCiot5G: false,
					N3Data: false, CPCiot5G: false, RestrictEC: 0x00, MCSI: false,
					EMCN3: false, EHCCPCiot5G: false, ATSIND: false, LCS5G: false,
				},
				T3512Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				T3502Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
			},
		},
		{
			name: "Case 2 - 5 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.PLMNList                          0x4A   TLV,    5-47B, 9.11.3.45
				0x4A, 0x06, 0x76, 0x54, 0x32, 0x98, 0x76, 0x54,
				// ie.RejectedNSSAI                     0x11   TLV,    4-42B, 9.11.3.46
				// ie.NSSAI                             0x31   TLV,   4-146B, 9.11.3.37
				0x31, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.PDUSessStatus                     0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0xaa,
				// ie.PDUSessReactivationResult         0x26   TLV,    4-34B, 9.11.3.42
				0x26, 0x02, 0xAA, 0x55,
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				EquivalentPlmns: &ie.PLMNList{
					PlmnIds: []ie.PlmnId{{
						MCC: "674",
						MNC: "235",
					}, {
						MCC: "896",
						MNC: "457",
					}},
				},
				// RejectedNSSAI   : &ie.RejectedNSSAI{},
				ConfiguredNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, true, false, true, false, true, false, true,
					}},
				},
				PDUSessReactivationResult: &ie.PDUSessReactivationResult{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, true, false, true, false, true,
						true, false, true, false, true, false, true, false,
					}},
				},
			},
		},
		{
			name: "Case 3 - 5 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.PDUSessReactivationResultErrCause 0x72 TLV-E,   5-515B, 9.11.3.43
				// ie.LADNInfo                          0x79 TLV-E, 12-1715B, 9.11.3.30
				// ie.MICOInd                           0xB0    TV,       1B, 9.11.3.31
				// ie.NwSlicingInd                      0x90    TV,       1B, 9.11.3.36
				0x92,
				// ie.SvcAreaList                       0x27   TLV,   6-114B, 9.11.3.49
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// PDUSessReactivationResultErrCause : &ie.PDUSessReactivationResultErrCause{},
				// LADNInfo                          : &ie.LADNInfo{},
				// MICOInd                           : &ie.MICOInd{},
				NwSlicingInd: &ie.NwSlicingInd{
					DCNI: true, NSSCI: false,
				},
				// SvcAreaList                       : &ie.SvcAreaList{},
			},
		},
		{
			name: "Case 4 - 5 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.GPRSTimer2                        0x5D   TLV,       3B, 9.11.2.4
				0x5D, 0x01, 0x2c,
				// ie.EmergNumList                        0x34   TLV,    5-50B, 9.11.3.23
				// ie.ExtendedEmergNumList                0x7A TLV-E, 7-65538B, 9.11.3.26
				// ie.SORTransparentCntr                         0x73 TLV-E,    20-nB, 9.11.3.51
				// ie.EAPMsg                            0x78 TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				Non3GppDeregTimerValue: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_1Min,
					Value: 12,
				},
				// EmergNumList                       : &ie.EmergNumList{},
				// ExtendedEmergNumList               : &ie.ExtendedEmergNumList{},
				// SORTransparentCntr                        : &ie.SORTransparentCntr{},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
			},
		},
		{
			name: "Case 5 - 5 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.ExtendedDRXParams                 0x6E   TLV,       3B, 9.11.3.26A
				// ie.GPRSTimer3                        0x6C   TLV,       3B, 9.11.2.5
				0x6C, 0x01, 0x06,
				// ie.GPRSTimer2                        0x6B   TLV,       3B, 9.11.2.4
				0x6B, 0x01, 0x56,
				// Rejected NSSAI
				// ie.GPRSTimer3                        0x6A   TLV,       3B, 9.11.2.5
				0x6A, 0x01, 0x06,
				// ie.UERadioCapabilityID               0x67   TLV,     3-nB, 9.11.3.68
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// NegotiatedExtendedDRXParams      : &ie.ExtendedDRXParams{},
				T3447Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				T3448Value: &ie.GPRSTimer2{
					Unit:  ie.TimerIncIn_Decihours,
					Value: 0x16,
				},
				T3324Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				// UERadioCapabilityID              : &ie.UERadioCapabilityID{},
			},
		},
		{
			name: "Case 6 - 5 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.UERadioCapabilityIDDelInd         0xE0    TV,       1B, 9.11.3.69
				// ie.NSSAI                             0x39   TLV,   4-146B, 9.11.3.37
				0x39, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// ie.CipheringKeyData                  0x74 TLV-E,    34-nB, 9.11.3.18C
				// ie.CAGInfoList                       0x75 TLV-E,     3-nB, 9.11.3.18A
				// ie.Truncated5GSTMSICfg               0x1B   TLV,       3B, 9.11.3.70
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// UERadioCapabilityIDDelInd        : &ie.UERadioCapabilityIDDelInd{},
				PendingNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				// CipheringKeyData                 : &ie.CipheringKeyData{},
				// CAGInfoList                      : &ie.CAGInfoList{},
				// Truncated5GSTMSICfg              : &ie.Truncated5GSTMSICfg{},
			},
		},
		{
			name: "Case 7 - 2 Optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x42, // EPD, SecHdr/Spare, Msg Type
				// ie.RegResult5GS,     LV,       2B, 9.11.3.6
				0x01, 0x01,
				// ie.WUSAssistanceInfo                 0x1C   TLV,     3-nB, 9.11.3.71
				// ie.NBN1ModeDRXParams                 0x29   TLV,       3B, 9.11.3.73
			},
			input: &RegAccept{
				RegResult5GS: &ie.RegResult5GS{
					EmergRegistered: false, NSSAAPerformed: false,
					SMSAllowed: false, Value: 1,
				},
				// NegotiatedWUSAssistanceInfo      : &ie.WUSAssistanceInfo{},
				// NegotiatedNBN1ModeDRXParams      : &ie.NBN1ModeDRXParams{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}
