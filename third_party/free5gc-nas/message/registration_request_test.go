package message

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestRegReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *RegReq
	}{
		{
			name: "case 1 - test data from wireshark",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x35,
				0x01, 0x02, 0xf8, 0x39, 0x0f, 0xff, 0x01, 0x01,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71,
				0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef,
				0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d,
				0x14, 0x20, 0x0a, 0x1d, 0x8a,
				// ie.UESecCapability TLV,    4-10B, 9.11.3.54
				// ie.UESecCapability 0x2E   TLV,    4-10B, 9.11.3.54
				0x2e, 0x04, 0x80, 0x20, 0xe0, 0xe0,
				// ie.S1UENwCapability 0x17   TLV, 4-15B, 9.11.3.48
				0x17, 0x07,
				0xe0, 0xe0, 0xc0, 0x40, 0x00, 0x80, 0x20,
				// ie.UEStatus         0x2B   TLV,       3B, 9.11.3.56
				0x2b, 0x01, 0x00,
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId:           ie.IdType_5GS_SUCI,
					SUPIFormat:         0x00,
					ProtectionSchemeId: ie.ECIESSchemeProfileA,
					HomeNwPubKeyId:     1,
					PlmnId: ie.PlmnId{
						MCC: "208",
						MNC: "93",
					},
					RoutingIndDigit: [4]uint8{15, 0, 15, 15},
					ECCEphPubKey: []byte{
						0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71,
						0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
						0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef,
						0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
					},
					CipherVal: []byte{0xb8, 0x6f, 0xb0, 0xfe, 0x25},
					MACTag: []byte{
						0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a,
					},
				}, // MobileId5GS
				UESecCapability: &ie.UESecCapability{
					Length: 4,
					EA05G:  true, EA1_128_5G: false, EA2_128_5G: false, EA3_128_5G: false,
					EA45G: false, EA55G: false, EA65G: false, EA75G: false,
					IA05G: false, IA1_128_5G: false, IA2_128_5G: true, IA3_128_5G: false,
					IA45G: false, IA55G: false, IA65G: false, IA75G: false,
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
				}, // UESecCapability
				S1UENwCapability: &ie.S1UENwCapability{
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
					UEA0: true, UEA1: true, UEA2: false, UEA3: false,
					UEA4: false, UEA5: false, UEA6: false, UEA7: false,
					UCS2: false, UIA1: true, UIA2: false, UIA3: false,
					UIA4: false, UIA5: false, UIA6: false, UIA7: false,
					ProSe_dd: false, ProSe: false, H245_ASH: false, ACC_CSFB: false,
					LPP: false, LCS: false, SRVCC_1x: false, NF: false,
					EPCO: true, HC_CP_CIoT: false, ERw_oPDN: false, S1U_Data: false,
					UP_CIoT: false, CP_CIoT: false, ProseRelay: false, ProSe_dc: false,
					Bearers15: false, SGC: false, N1mode: true, DCNR: false,
					CP_Backoff: false, RestrictEC: false, V2X_PC5: false, MultipleDRB: false,
					V2X_NR_PC5: false, UP_MT_EDT: false, CP_MT_EDT: false, WUSA: false,
					RACS: false,
				}, // S1UENwCapability
				UEStatus: &ie.UEStatus{N1ModeReg: false, S1ModeReg: false},
			},
		},
		{
			name: "case 2 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// NoncurrentNativeNASKeySetId: &ie.NASKeySetId 0xC0    TV,       1B, 9.11.3.32
				0xC9,
				// Capability5GMM                     : &ie.Capability5GMM      0x10   TLV,    3-15B, 9.11.3.1
				0x10, 0x03, 0xaa, 0xaa, 0x0a,
				// ReqNSSAI                           : &ie.NSSAI               0x2F   TLV,    4-74B, 9.11.3.37
				0x2F, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// LastVisitedRegisteredTAI           : &ie.TrackingAreaId5GS       0x52    TV,       7B, 9.11.3.8
				0x52, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
				// UplinkDataStatus                   : &ie.UplinkDataStatus    0x40   TLV,    4-34B, 9.11.3.57
				0x40, 0x02, 0x76, 0x84,
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				NoncurrentNativeNASKeySetId: &ie.NASKeySetId{
					Tsc: 1, Ksi: 1,
				},
				Capability5GMM: &ie.Capability5GMM{
					Length: 3,
					SGC:    true, IPHCCPCiot5G: false, N3Data: false, CPCiot5G: false,
					RestrictEC: true, LPP: false, HOAttach: true, S1Mode: false,
					RACS: true, NSSAA: false, LCS5G: true, V2XCNPC5: false,
					V2XCEPC5: true, V2X: false, UPCiot5G: true, SRVCC5G: false,
					EHCCPCiot5G: true, Multipleup: false, WUSA: true, CAG: false,
				},
				ReqNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				LastVisitedRegisteredTAI: &ie.TrackingAreaId5GS{
					PlmnId: ie.PlmnId{
						MCC: "310",
						MNC: "310",
					},
					TAC: "000001",
				},
				UplinkDataStatus: &ie.UplinkDataStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
			},
		},
		{
			name: "case 3 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// PDUSessStatus        : &ie.PDUSessStatus        0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0x00,
				// MICOInd              : &ie.MICOInd              0xB0    TV,       1B, 9.11.3.31
				// AdditionalGUTI       : &ie.MobileId5GS          0x77 TLV-E,      14B, 9.11.3.4
				0x77, 0x00, 0x0b,
				0xf2, 0x02, 0xf8, 0x39, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01,
				// AllowedPDUSessStatus : &ie.AllowedPDUSessStatus 0x25   TLV,    4-34B, 9.11.3.13
				0x25, 0x02, 0x76, 0x84,
				// UesUsageSetting      : &ie.UesUsageSetting      0x18   TLV,       3B, 9.11.3.55
				0x18, 0x01, 0x76,
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				AdditionalGUTI: &ie.MobileId5GS{
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
				AllowedPDUSessStatus: &ie.AllowedPDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				UesUsageSetting: &ie.UesUsageSetting{
					UESUsageSetting: uint8(ie.UsageType_VoiceCentric),
				},
			},
		},
		{
			name: "case 4 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// ReqDRXParams    : &ie.DRXParams5GS    0x51   TLV,       3B, 9.11.3.2A
				// EPSNASMsgCntr   : &ie.EPSNASMsgCntr   0x70 TLV-E,     4-nB, 9.11.3.24
				// LADNInd         : &ie.LADNInd         0x74 TLV-E,   3-811B, 9.11.3.29
				0x74, 0x00, 0x05,
				0x04, 0x03, 0x61, 0x62, 0x63,
				// PayloadCntrType : &ie.PayloadCntrType 0x80    TV,       1B, 9.11.3.40
				0x81,
				// PayloadCntr     : &ie.PayloadCntr     0x7B TLV-E, 4-65538B, 9.11.3.39
				0x7b, 0x00, 0x08,
				0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1,
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				// ReqDRXParams   : &ie.DRXParams5GS  0x51   TLV,       3B, 9.11.3.2A
				// EPSNASMsgCntr  : &ie.EPSNASMsgCntr 0x70 TLV-E,     4-nB, 9.11.3.24
				LADNInd: &ie.LADNInd{
					DNNs: []ie.DNN{{
						Value: "abc",
					}},
				},
				PayloadCntrType: &ie.PayloadCntrType{
					Value: ie.PayloadCntrType_N1SMInfo,
				},
				PayloadCntr: &ie.PayloadCntr{
					Pct:      ie.PayloadCntrType_N1SMInfo,
					Contents: []byte{0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
				},
			},
		},
		{
			name: "case 5 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// NwSlicingInd            : &ie.NwSlicingInd            0x90    TV,       1B, 9.11.3.36
				0x92,
				// UpdateType5GS              : &ie.UpdateType5GS              0x53   TLV,       3B, 9.11.3.9A
				0x53, 0x01, 0x76,
				// MobileStationClassmark2 : &ie.MobileStationClassmark2 0x41   TLV,       5B, 9.11.3.31C
				// SupportedCodecs               : &ie.SupportedCodecList            0x42   TLV,     5-nB, 9.11.3.51A
				// NASMsgCntr              : &ie.NASMsgCntr              0x71 TLV-E,     4-nB, 9.11.3.33
				0x71, 0x00, 0x01, 0x43,
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				NwSlicingInd: &ie.NwSlicingInd{
					DCNI: true, NSSCI: false,
				},
				UpdateType5GS: &ie.UpdateType5GS{
					EPS_PNB_CIoT: 0x03, PNB_CIoT_5GS: 0x01,
					NG_RAN_RCU: true, SMSRequested: false,
				},
				NASMsgCntr: &ie.NASMsgCntr{
					Contents: []byte{0x43},
				},
			},
		},
		{
			name: "case 6 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// EPSBearerCtxStatus   : &ie.EPSBearerCtxStatus  0x60   TLV,    4B, 9.11.3.23A
				// ReqExtendedDRXParams : &ie.ExtendedDRXParams   0x6E   TLV,    3B, 9.11.3.26A
				// T3324Value           : &ie.GPRSTimer3          0x6A   TLV,    3B, 9.11.2.5
				0x6A, 0x01, 0x06,
				// UERadioCapabilityID  : &ie.UERadioCapabilityID 0x67   TLV,  3-nB, 9.11.3.68
				// ReqMappedNSSAI       : &ie.MappedNSSAI         0x35   TLV, 3-42B, 9.11.3.31B
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				T3324Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
			},
		},
		{
			name: "case 7 another 5 optional IEs",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// AdditionalInfoReq    : &ie.AdditionalInfoReq 0x48   TLV,   3B, 9.11.3.12A
				// ReqWUSAssistanceInfo : &ie.WUSAssistanceInfo 0x1A   TLV, 3-nB, 9.11.3.71
				// N5GCInd              : &ie.N5GCInd           0xA0     T,   1B, 9.11.3.72
				// ReqNBN1ModeDRXParams : &ie.NBN1ModeDRXParams 0x30   TLV,   3B, 9.11.3.73
			},
			expected: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(RegReq)
			err := ie.UnmarshalBinary(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRegReqUnmarshalBinaryNegative(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected error
	}{
		{
			// https://github.com/free5gc/nas/pull/24
			name: "case 1 - S1 S1UENwCapability bytes len not match",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x35,
				0x01, 0x02, 0xf8, 0x39, 0x0f, 0xff, 0x01, 0x01,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71,
				0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef,
				0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d,
				0x14, 0x20, 0x0a, 0x1d, 0x8a,
				// ie.UESecCapability TLV,    4-10B, 9.11.3.54
				// ie.UESecCapability 0x2E   TLV,    4-10B, 9.11.3.54
				0x2e, 0x04, 0x80, 0x20, 0xe0, 0xe0,
				// ie.S1UENwCapability 0x17   TLV, 4-15B, 9.11.3.48
				0x17, 0x08,
				0xe0, 0xe0, 0xc0, 0x40, 0x00, 0x80, // 0x20, 0x00,
			},
			expected: errors.Errorf(
				"RegReq UnmarshalBinary getIeLen of S1UENwCapability: IE length 8 exceeds remaining data 6"),
		},
		{
			name: "case 2 - RegReqCntr has no container type decoded before container value",
			input: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// ReqDRXParams    : &ie.DRXParams5GS    0x51   TLV,       3B, 9.11.3.2A
				// EPSNASMsgCntr   : &ie.EPSNASMsgCntr   0x70 TLV-E,     4-nB, 9.11.3.24
				// LADNInd         : &ie.LADNInd         0x74 TLV-E,   3-811B, 9.11.3.29
				0x74, 0x00, 0x05,
				0x04, 0x03, 0x61, 0x62, 0x63,
				// PayloadCntrType : &ie.PayloadCntrType 0x80    TV,       1B, 9.11.3.40
				// 0x81,
				// PayloadCntr     : &ie.PayloadCntr     0x7B TLV-E, 4-65538B, 9.11.3.39
				0x7b, 0x00, 0x08,
				0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1,
			},
			expected: errors.Errorf(
				"RegReq UnmarshalBinary no PayloadCntrType for PayloadCntr"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(RegReq)
			err := ie.UnmarshalBinary(tc.input)
			require.EqualError(t, tc.expected, err.Error())
		})
	}
}

func TestRegReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *RegReq
		expected []byte
	}{
		{
			name: "case 1 - test data from wireshark",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x35,
				0x01, 0x02, 0xf8, 0x39, 0x0f, 0xff, 0x01, 0x01,
				0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71,
				0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
				0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef,
				0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
				0xb8, 0x6f, 0xb0, 0xfe, 0x25, 0x29, 0xd7, 0x2d,
				0x14, 0x20, 0x0a, 0x1d, 0x8a,
				// ie.UESecCapability TLV,    4-10B, 9.11.3.54
				// ie.UESecCapability 0x2E   TLV,    4-10B, 9.11.3.54
				0x2e, 0x04, 0x80, 0x20, 0xe0, 0xe0,
				// ie.S1UENwCapability 0x17   TLV, 4-15B, 9.11.3.48
				0x17, 0x08,
				0xe0, 0xe0, 0xc0, 0x40, 0x00, 0x80, 0x20, 0x00,
				// ie.UEStatus         0x2B   TLV,       3B, 9.11.3.56
				0x2b, 0x01, 0x00,
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId:           ie.IdType_5GS_SUCI,
					SUPIFormat:         0x00,
					ProtectionSchemeId: ie.ECIESSchemeProfileA,
					HomeNwPubKeyId:     1,
					PlmnId: ie.PlmnId{
						MCC: "208",
						MNC: "93",
					},
					RoutingIndDigit: [4]uint8{15, 0, 15, 15},
					ECCEphPubKey: []byte{
						0x2b, 0xc3, 0xc1, 0xba, 0xef, 0xf3, 0xcb, 0x71,
						0x58, 0xca, 0x0e, 0x35, 0x70, 0x35, 0x9f, 0xdf,
						0x9b, 0x11, 0xcc, 0xe0, 0x59, 0xa9, 0x0c, 0xef,
						0xef, 0xda, 0x6b, 0x9d, 0xa1, 0x7a, 0x5f, 0x0b,
					},
					CipherVal: []byte{0xb8, 0x6f, 0xb0, 0xfe, 0x25},
					MACTag: []byte{
						0x29, 0xd7, 0x2d, 0x14, 0x20, 0x0a, 0x1d, 0x8a,
					},
				}, // MobileId5GS
				UESecCapability: &ie.UESecCapability{
					Length: 4,
					EA05G:  true, EA1_128_5G: false, EA2_128_5G: false, EA3_128_5G: false,
					EA45G: false, EA55G: false, EA65G: false, EA75G: false,
					IA05G: false, IA1_128_5G: false, IA2_128_5G: true, IA3_128_5G: false,
					IA45G: false, IA55G: false, IA65G: false, IA75G: false,
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
				}, // UESecCapability
				S1UENwCapability: &ie.S1UENwCapability{
					EEA0: true, EEA1_128: true, EEA2_128: true, EEA3_128: false,
					EEA4: false, EEA5: false, EEA6: false, EEA7: false,
					EIA0: true, EIA1_128: true, EIA2_128: true, EIA3_128: false,
					EIA4: false, EIA5: false, EIA6: false, EIA7: false,
					UEA0: true, UEA1: true, UEA2: false, UEA3: false,
					UEA4: false, UEA5: false, UEA6: false, UEA7: false,
					UCS2: false, UIA1: true, UIA2: false, UIA3: false,
					UIA4: false, UIA5: false, UIA6: false, UIA7: false,
					ProSe_dd: false, ProSe: false, H245_ASH: false, ACC_CSFB: false,
					LPP: false, LCS: false, SRVCC_1x: false, NF: false,
					EPCO: true, HC_CP_CIoT: false, ERw_oPDN: false, S1U_Data: false,
					UP_CIoT: false, CP_CIoT: false, ProseRelay: false, ProSe_dc: false,
					Bearers15: false, SGC: false, N1mode: true, DCNR: false,
					CP_Backoff: false, RestrictEC: false, V2X_PC5: false, MultipleDRB: false,
					V2X_NR_PC5: false, UP_MT_EDT: false, CP_MT_EDT: false, WUSA: false,
					RACS: false,
				}, // S1UENwCapability
				UEStatus: &ie.UEStatus{N1ModeReg: false, S1ModeReg: false},
			},
		},
		{
			name: "case 2 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// NoncurrentNativeNASKeySetId: &ie.NASKeySetId 0xC0    TV,       1B, 9.11.3.32
				0xC9,
				// Capability5GMM                     : &ie.Capability5GMM      0x10   TLV,    3-15B, 9.11.3.1
				0x10, 0x03, 0xaa, 0xaa, 0x0a,
				// ReqNSSAI                           : &ie.NSSAI               0x2F   TLV,    4-74B, 9.11.3.37
				0x2F, 0x0a,
				0x04, 0x01, 0x01, 0x02, 0x03, 0x04, 0x01, 0x11, 0x22, 0x33,
				// LastVisitedRegisteredTAI           : &ie.TrackingAreaId5GS       0x52    TV,       7B, 9.11.3.8
				0x52, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
				// UplinkDataStatus                   : &ie.UplinkDataStatus    0x40   TLV,    4-34B, 9.11.3.57
				0x40, 0x02, 0x76, 0x84,
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				NoncurrentNativeNASKeySetId: &ie.NASKeySetId{
					Tsc: 1, Ksi: 1,
				},
				Capability5GMM: &ie.Capability5GMM{
					Length: 3,
					SGC:    true, IPHCCPCiot5G: false, N3Data: false, CPCiot5G: false,
					RestrictEC: true, LPP: false, HOAttach: true, S1Mode: false,
					RACS: true, NSSAA: false, LCS5G: true, V2XCNPC5: false,
					V2XCEPC5: true, V2X: false, UPCiot5G: true, SRVCC5G: false,
					EHCCPCiot5G: true, Multipleup: false, WUSA: true, CAG: false,
				},
				ReqNSSAI: &ie.NSSAI{
					SNSSAIs: []ie.SNSSAI{{
						SST: 0x01, SD: "010203",
					}, {
						SST: 0x01, SD: "112233",
					}},
				},
				LastVisitedRegisteredTAI: &ie.TrackingAreaId5GS{
					PlmnId: ie.PlmnId{
						MCC: "310",
						MNC: "310",
					},
					TAC: "000001",
				},
				UplinkDataStatus: &ie.UplinkDataStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
			},
		},
		{
			name: "case 3 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// PDUSessStatus        : &ie.PDUSessStatus        0x50   TLV,    4-34B, 9.11.3.44
				0x50, 0x02, 0x02, 0x00,
				// MICOInd              : &ie.MICOInd              0xB0    TV,       1B, 9.11.3.31
				// AdditionalGUTI       : &ie.MobileId5GS          0x77 TLV-E,      14B, 9.11.3.4
				0x77, 0x00, 0x0b,
				0xf2, 0x02, 0xf8, 0x39, 0xca, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x01,
				// AllowedPDUSessStatus : &ie.AllowedPDUSessStatus 0x25   TLV,    4-34B, 9.11.3.13
				0x25, 0x02, 0x76, 0x84,
				// UesUsageSetting      : &ie.UesUsageSetting      0x18   TLV,       3B, 9.11.3.55
				0x18, 0x01, 0x01,
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				PDUSessStatus: &ie.PDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					}},
				},
				AdditionalGUTI: &ie.MobileId5GS{
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
				AllowedPDUSessStatus: &ie.AllowedPDUSessStatus{
					Psi: ie.Psi{PSI: [16]bool{
						false, true, true, false, true, true, true, false,
						false, false, true, false, false, false, false, true,
					}},
				},
				UesUsageSetting: &ie.UesUsageSetting{
					UESUsageSetting: uint8(ie.UsageType_DataCentric),
				},
			},
		},
		{
			name: "case 4 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// ReqDRXParams    : &ie.DRXParams5GS    0x51   TLV,       3B, 9.11.3.2A
				// EPSNASMsgCntr   : &ie.EPSNASMsgCntr   0x70 TLV-E,     4-nB, 9.11.3.24
				// LADNInd         : &ie.LADNInd         0x74 TLV-E,   3-811B, 9.11.3.29
				0x74, 0x00, 0x05,
				0x04, 0x03, 0x61, 0x62, 0x63,
				// PayloadCntrType : &ie.PayloadCntrType 0x80    TV,       1B, 9.11.3.40
				0x81,
				// PayloadCntr     : &ie.PayloadCntr     0x7B TLV-E, 4-65538B, 9.11.3.39
				0x7b, 0x00, 0x08,
				0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1,
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				// ReqDRXParams   : &ie.DRXParams5GS  0x51   TLV,       3B, 9.11.3.2A
				// EPSNASMsgCntr  : &ie.EPSNASMsgCntr 0x70 TLV-E,     4-nB, 9.11.3.24
				LADNInd: &ie.LADNInd{
					DNNs: []ie.DNN{{
						Value: "abc",
					}},
				},
				PayloadCntrType: &ie.PayloadCntrType{
					Value: ie.PayloadCntrType_N1SMInfo,
				},
				PayloadCntr: &ie.PayloadCntr{
					Pct:      ie.PayloadCntrType_N1SMInfo,
					Contents: []byte{0x2e, 0x05, 0x01, 0xc1, 0xff, 0xff, 0x91, 0xa1},
				},
			},
		},
		{
			name: "case 5 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// NwSlicingInd            : &ie.NwSlicingInd            0x90    TV,       1B, 9.11.3.36
				0x92,
				// UpdateType5GS              : &ie.UpdateType5GS              0x53   TLV,       3B, 9.11.3.9A
				0x53, 0x01, 0x36,
				// MobileStationClassmark2 : &ie.MobileStationClassmark2 0x41   TLV,       5B, 9.11.3.31C
				// SupportedCodecs               : &ie.SupportedCodecList            0x42   TLV,     5-nB, 9.11.3.51A
				// NASMsgCntr              : &ie.NASMsgCntr              0x71 TLV-E,     4-nB, 9.11.3.33
				0x71, 0x00, 0x01, 0x43,
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				NwSlicingInd: &ie.NwSlicingInd{
					DCNI: true, NSSCI: false,
				},
				UpdateType5GS: &ie.UpdateType5GS{
					EPS_PNB_CIoT: 0x03, PNB_CIoT_5GS: 0x01,
					NG_RAN_RCU: true, SMSRequested: false,
				},
				NASMsgCntr: &ie.NASMsgCntr{
					Contents: []byte{0x43},
				},
			},
		},
		{
			name: "case 6 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// EPSBearerCtxStatus   : &ie.EPSBearerCtxStatus  0x60   TLV,    4B, 9.11.3.23A
				// ReqExtendedDRXParams : &ie.ExtendedDRXParams   0x6E   TLV,    3B, 9.11.3.26A
				// T3324Value           : &ie.GPRSTimer3          0x6A   TLV,    3B, 9.11.2.5
				0x6A, 0x01, 0x06,
				// UERadioCapabilityID  : &ie.UERadioCapabilityID 0x67   TLV,  3-nB, 9.11.3.68
				// ReqMappedNSSAI       : &ie.MappedNSSAI         0x35   TLV, 3-42B, 9.11.3.31B
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
				T3324Value: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
			},
		},
		{
			name: "case 7 another 5 optional IEs",
			expected: []byte{
				0x7e, 0x00, 0x41, // EPD, SecHdr/Spare, Msg Type
				// ie.RegType5GS + ie.NASKeySetId, V, 1/2B
				0x79,
				// ie.MobileId5GS  LV-E,     6-nB, 9.11.3.4
				0x00, 0x09,
				0x07, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				// AdditionalInfoReq    : &ie.AdditionalInfoReq 0x48   TLV,   3B, 9.11.3.12A
				// ReqWUSAssistanceInfo : &ie.WUSAssistanceInfo 0x1A   TLV, 3-nB, 9.11.3.71
				// N5GCInd              : &ie.N5GCInd           0xA0     T,   1B, 9.11.3.72
				// ReqNBN1ModeDRXParams : &ie.NBN1ModeDRXParams 0x30   TLV,   3B, 9.11.3.73
			},
			input: &RegReq{
				RegType5GS: &ie.RegType5GS{FOR_Pending: true, Value: 1},
				Ngksi:      &ie.NASKeySetId{Tsc: 0, Ksi: 7},
				MobileId5GS: &ie.MobileId5GS{
					TypeOfId: ie.IdType_5GS_EUI64,
					EUI64:    [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
				}, // MobileId5GS
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}
