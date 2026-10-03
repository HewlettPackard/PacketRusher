package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessEstAcceptUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
		toDoIEs  bool
	}{
		{
			name: "Real packet 001",
			input: []byte{
				0x2e, 0x0a, 0x00, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x0f,
				0x01, 0x00, 0x06, // QoS rule id + length
				0x31,             // opcode: 1, DQR: 1, Num of pkt filter: 1
				0x31, 0x01, 0x01, // 0x31:(Bidir, id=1); 0x01:len; 0x01: Match-all type;
				0xff, 0x09, // QoS Precedence : 0xff, Segregation: 0, QFI: 09
				0x02, 0x00, 0x03, // QoS rule id + length
				0x20,       // 0x20:(opCode: 1, DQR=0; num of pkt filter=0);
				0x17, 0x09, // QoS Precedence : 0x17, Segregation: 0, QFI: 09
				// ie.SessAMBR LV
				0x06,
				0x01, 0x03, 0xe8, 0x01, 0x03, 0xe8,
				// PDU address, TLV
				0x29, 0x05, 0x01, 0x0a, 0x3c, 0x00, 0x06,
				// S-NSSAI, TLV
				0x22, 0x04, 0x01, 0x01, 0x02, 0x03,
				// Authorized QoS flow descriptions, TLV-E
				0x79, 0x00, 0x0c,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				// Extended protocol configuration options, TLV-E
				0x7b, 0x00, 0x0d,
				0x80, 0x00, 0x0d, 0x04, 0x08, 0x08, 0x08, 0x08, 0x00, 0x10, 0x02, 0x05, 0x4e,
				// DNN, TLV
				0x25, 0x09, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74,
			},
			expected: &PDUSessEstAccept{
				PDUSessId: 10,
				PTI:       0,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{{
						RuleId:       1,
						OpCode:       ie.OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []ie.PacketFilter{{
							Dir: ie.PFD_BiDir,
							Id:  1,
							Contents: ie.PacketFilterContents{
								MatchAll:   true,
								RemoteAddr: "any",
								LocalAddr:  "assigned",
							},
						}},
					}, {
						RuleId:        2,
						OpCode:        ie.OpCode_CreateNewQosRule,
						IsDefaultDQR:  false,
						Precedence:    23,
						Segregation:   false,
						QFI:           9,
						PktFilterList: []ie.PacketFilter{},
					}},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Kbps,
					ValueDownlink: 1000,
					UnitUplink:    ie.Rate_1Kbps,
					ValueUplink:   1000,
				},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
					},
				},
				PDUAddr: &ie.PDUAddr{
					IPv4: []byte{0x0a, 0x3c, 0x00, 0x06},
				},
				SNSSAI: &ie.SNSSAI{
					SST: 0x01,
					SD:  "010203",
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1358,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
			},
		},
		// Negative Cases
		// ====================================================
		{
			name:  "length = 0, < 4",
			input: []byte{
				// EPD, PDUSessId, PTI, MsgType
			},
		},
		{
			name: "length = 1,  < 4",
			input: []byte{
				0x2e, // EPD, PDUSessId, PTI, MsgType
			},
		},
		{
			name: "length = 2,  < 4",
			input: []byte{
				0x2e, 0x05, // EPD, PDUSessId, PTI, MsgType
			},
		},
		{
			name: "length = 3,  < 4",
			input: []byte{
				0x2e, 0x05, 0x01, // EPD, PDUSessId, PTI, MsgType
			},
		},
		{
			name: "length = 4, w/o mandatory IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
			},
		},
		{
			name: "for 2 bytes length IEs (LV-E), only 1 byte left in the buffer",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x09,
			},
		},
		{
			name: "LV-E IE, only length part in the buffer",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
			},
		},
		{
			name: "LV-E IE's len (9) > len(5) left in the buffer",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31,
			},
		},
		{
			name: "Mandatory IE not appear in order",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.SessAMBR LV
				0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
			},
		},
		{
			name:    "With unimplemented Optional IE",
			toDoIEs: true,
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.SessAMBR LV
				0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// ie.AlwaysonPDUSessInd , TV, 1B (PacketRusher compatibility codec)
				0x80,
				// ie.MappedEPSBearerCtxs, TLV-E, 7-65538B (PacketRusher compatibility codec)
				0x75, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ie.ServingPLMNRateCtrl, TLV, 4B (unimplemented)
				0x18, 0x02,
				0x01, 0x02,
				// ie.ATSSSCntr, TLV-E, 3-65538B (unimplemented)
				0x77, 0x00, 0x00,
				// ie.CtrlPlaneOnlyInd, TV, 1B (unimplemented)
				0xC0,
				// ie.IPHdrCompressionCfg, TLV, 5-257B (unimplemented)
				0x66, 0x03,
				0x00, 0x00, 0x00,
				// ie.EthHdrCompressionCfg, TLV, 3B (unimplemented)
				0x1f, 0x01,
				0x00,
			},
			// Unimplemented IE ignored, implemented parts are there
			expected: &PDUSessEstAccept{
				AlwaysonPDUSessInd:  &ie.AlwaysonPDUSessInd{APSI: false},
				MappedEPSBearerCtxs: &ie.MappedEPSBearerCtxs{Contents: []byte{1, 2, 3, 4}},
				PDUSessId:           5,
				PTI:                 1,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{
						{
							RuleId:       1,
							OpCode:       ie.OpCode_CreateNewQosRule,
							IsDefaultDQR: true,
							Precedence:   255,
							Segregation:  false,
							QFI:          9,
							PktFilterList: []ie.PacketFilter{
								{
									Dir: ie.PFD_BiDir,
									Id:  1,
									Contents: ie.PacketFilterContents{
										MatchAll:   true,
										RemoteAddr: "any",
										LocalAddr:  "assigned",
									},
								},
							},
						},
					},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Mbps,
					ValueDownlink: 100,
					UnitUplink:    ie.Rate_1Mbps,
					ValueUplink:   200,
				},
			},
		},
		{
			// shall handle only the contents of the information element appearing
			// first and shall ignore all subsequent repetitions of the IE.
			name: "With repeated IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.SessAMBR LV
				0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// ie.NwFeatureSupport5GSM TLV
				0x17, 0x01, 0x01,
				// ie.NwFeatureSupport5GSM TLV (Repeated)
				0x17, 0x01, 0x02,
				// ie.NwFeatureSupport5GSM TLV (Repeated)
				0x17, 0x01, 0x03,
				// ie.NwFeatureSupport5GSM TLV (Repeated)
				0x17, 0x01, 0x04,
			},
			// Unimplemented IE ignored, implemented parts are there
			expected: &PDUSessEstAccept{
				PDUSessId: 5,
				PTI:       1,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{
						{
							RuleId:       1,
							OpCode:       ie.OpCode_CreateNewQosRule,
							IsDefaultDQR: true,
							Precedence:   255,
							Segregation:  false,
							QFI:          9,
							PktFilterList: []ie.PacketFilter{
								{
									Dir: ie.PFD_BiDir,
									Id:  1,
									Contents: ie.PacketFilterContents{
										MatchAll:   true,
										RemoteAddr: "any",
										LocalAddr:  "assigned",
									},
								},
							},
						},
					},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Mbps,
					ValueDownlink: 100,
					UnitUplink:    ie.Rate_1Mbps,
					ValueUplink:   200,
				},
				NwFeatureSupport5GSM: &ie.NwFeatureSupport5GSM{
					EPTS1: 1,
				},
			},
		},

		// Positive Cases
		// ====================================================
		{
			name: "w/o optional IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.SessAMBR LV
				0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
			},
			expected: &PDUSessEstAccept{
				PDUSessId: 5,
				PTI:       1,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{
						{
							RuleId:       1,
							OpCode:       ie.OpCode_CreateNewQosRule,
							IsDefaultDQR: true,
							Precedence:   255,
							Segregation:  false,
							QFI:          9,
							PktFilterList: []ie.PacketFilter{
								{
									Dir: ie.PFD_BiDir,
									Id:  1,
									Contents: ie.PacketFilterContents{
										MatchAll:   true,
										RemoteAddr: "any",
										LocalAddr:  "assigned",
									},
								},
							},
						},
					},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Mbps,
					ValueDownlink: 100,
					UnitUplink:    ie.Rate_1Mbps,
					ValueUplink:   200,
				},
			},
		},
		{
			name: "Positive Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc2, // EPD, PDUSessId, PTI, MsgType
				0x11, // SelectedPDUSessType / SelectedSSCMode v
				// ie.QosRules LV-E
				0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.SessAMBR LV
				0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// ie.PDUAddr TLV
				0x29, 0x05, 0x01, 0x0a, 0x3c, 0x00, 0x01,
				// ie.SNSSAI TLV
				0x22, 0x04, 0x01, 0x01, 0x02, 0x03,
				// ie.QosFlowDescs TLV-E
				0x79, 0x00, 0x06,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				// ie.DNN TLV
				0x25, 0x09,
				0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74,
				// ie.Cause5GSM TV
				0x59, 0x00,
				// ie.RQTimerValue TV
				// ie.AlwaysonPDUSessInd TV
				// ie.MappedEPSBearerCtxs TLV-E
				// ie.EAPMsg TLV-E
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ie.ExtendedProtCfgOpts TLV-E
				0x7B, 0x00, 0x0D,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4, 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
				// ie.NwFeatureSupport5GSM TLV
				0x17, 0x01, 0x01,
				// ie.ServingPLMNRateCtrl TLV
				// ie.ATSSSCntr TLV-E
				// ie.CtrlPlaneOnlyInd TV
				// ie.IPHdrCompressionCfg TLV
				// ie.EthHdrCompressionCfg TLV
			},
			expected: &PDUSessEstAccept{
				PDUSessId: 5,
				PTI:       1,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{
						{
							RuleId:       1,
							OpCode:       ie.OpCode_CreateNewQosRule,
							IsDefaultDQR: true,
							Precedence:   255,
							Segregation:  false,
							QFI:          9,
							PktFilterList: []ie.PacketFilter{
								{
									Dir: ie.PFD_BiDir,
									Id:  1,
									Contents: ie.PacketFilterContents{
										MatchAll:   true,
										RemoteAddr: "any",
										LocalAddr:  "assigned",
									},
								},
							},
						},
					},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Mbps,
					ValueDownlink: 100,
					UnitUplink:    ie.Rate_1Mbps,
					ValueUplink:   200,
				},
				PDUAddr: &ie.PDUAddr{
					IPv4: []byte{0x0a, 0x3c, 0x00, 0x01},
				},
				SNSSAI: &ie.SNSSAI{
					SST: 1,
					SD:  "010203",
				},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
					},
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1400,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
				NwFeatureSupport5GSM: &ie.NwFeatureSupport5GSM{
					EPTS1: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessEstAccept)
			err := msg.UnmarshalBinary(tc.input)
			if tc.expected == nil {
				require.Error(t, err)
				_, isIEToDoErr := err.(*ie.IEToDo)
				require.True(t, !isIEToDoErr)
			} else {
				if tc.toDoIEs {
					require.Error(t, err)
					_, isIEToDoErr := err.(*Error)
					// fmt.Printf("%T, %v, %v", err, err, isIEToDoErr)
					require.Equal(t, tc.toDoIEs, isIEToDoErr)
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, tc.expected, msg)
			}
		})
	}
}

func TestPDUSessEstAcceptMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xc2, 0x11, 0x00, 0x09, 0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				0x06, 0x06, 0x00, 0x64, 0x06, 0x00, 0xc8, 0x29, 0x05, 0x01, 0x0a, 0x3c, 0x00, 0x01, 0x22, 0x04,
				0x01, 0x01, 0x02, 0x03, 0x79, 0x00, 0x06, 0x09, 0x20, 0x41, 0x01, 0x01, 0x09, 0x25, 0x09, 0x08,
				0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74,
			},
			input: &PDUSessEstAccept{
				PDUSessId: 5,
				PTI:       1,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{
						{
							RuleId:       1,
							OpCode:       ie.OpCode_CreateNewQosRule,
							IsDefaultDQR: true,
							Precedence:   255,
							Segregation:  false,
							QFI:          9,
							PktFilterList: []ie.PacketFilter{
								{
									Dir: ie.PFD_BiDir,
									Id:  1,
									Contents: ie.PacketFilterContents{
										MatchAll:   true,
										RemoteAddr: "any",
										LocalAddr:  "assigned",
									},
								},
							},
						},
					},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Mbps,
					ValueDownlink: 100,
					UnitUplink:    ie.Rate_1Mbps,
					ValueUplink:   200,
				},
				PDUAddr: &ie.PDUAddr{
					IPv4: []byte{0x0a, 0x3c, 0x00, 0x01},
				},
				SNSSAI: &ie.SNSSAI{
					SST: 1,
					SD:  "010203",
				},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
					},
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
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
