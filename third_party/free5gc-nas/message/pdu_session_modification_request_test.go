package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessModReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// Capability5GSM : &ie.Capability5GSM             0x28   TLV,    3-15B, 9.11.4.1
				0x28, 0x01, 0xaa,
				// Cause5GSM      : &ie.Cause5GSM                  0x59    TV,       2B, 9.11.4.2
				0x59, 0x00,
				// MaxNumOfSupportedPktFilters : &ie.MaxNumOfSupportedPktFilters 0x55    TV,       3B, 9.11.4.9
				// AlwaysonPDUSessReq   : &ie.AlwaysonPDUSessReq   0xB0    TV,       1B, 9.11.4.4
				0xB0,
				// IntegrityProtectionMaxDataRate : &ie.IntegrityProtectionMaxDataRate 0x13    TV,       3B, 9.11.4.7
				0x13, 0xff, 0xff,
				// ReqQosRules          : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				// ReqQosFlowDescs      : &ie.QosFlowDescs         0x79 TLV-E, 6-65538B, 9.11.4.12
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x2d,
				0x80,
				0x80, 0x21, 0x10, // Internet Protocol Control Protocol, 16B
				0x01, 0x00, 0x00, 0x10, // 16B
				0x81, 0x06, 0x00, 0x00, 0x00, 0x00, // Primary DNS server IP addr, 6B, 0.0.0.0
				0x83, 0x06, 0x00, 0x00, 0x00, 0x00, // 2nd DNS server IP addr, 6B, 0.0.0.0
				0x00, 0x0d, 0x00, // DNS IPv4 Req, 0B
				0x00, 0x0a, 0x00, // IP Addr Allocation via NAS Signaling, 0B
				0x00, 0x05, 0x00, // MS Support of Nw Req Bearer Ctrl Indicator, 0B
				0x00, 0x10, 0x00, // IPv4 Link MTU Req, 0B
				0x00, 0x11, 0x00, // MS Support of Local Addr in TFT Indicator, 0B
				0x00, 0x17, 0x01, 0x01, // 3GPP PS data off UE status, 1B
				0x00, 0x23, 0x00, // QoS Rules w/ the len of 2B support indicator, 0B
				0x00, 0x24, 0x00, // QoS flow desc w/ the len of 2B support indicator, 0B
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			expected: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
				},
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: false,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq:        true,
						DNSV4Req:              true,
						UEStatus3GPPPSDataOff: ie.UEStatus3GPPPSDataOff_Deactivate,
					},
				},
			},
		},
		{
			name: "Case 2 - another 5 optional IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// MaxNumOfSupportedPktFilters : &ie.MaxNumOfSupportedPktFilters 0x55    TV,       3B, 9.11.4.9
				// AlwaysonPDUSessReq   : &ie.AlwaysonPDUSessReq   0xB0    TV,       1B, 9.11.4.4
				// ReqQosRules          : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				0x7A, 0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ReqQosFlowDescs      : &ie.QosFlowDescs         0x79 TLV-E, 6-65538B, 9.11.4.12
				0x79, 0x00, 0x06,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
			},
			expected: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
				ReqQosRules: &ie.QosRules{
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
					}},
				},
				ReqQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{{
						QFI:    9,
						OpCode: ie.QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(ie.Qfd_5QI9),
					}},
				},
			},
		},
		{
			name: "Case 3 - another 3 optional IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			expected: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessModReq)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessModReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// Capability5GSM : &ie.Capability5GSM             0x28   TLV,    3-15B, 9.11.4.1
				0x28, 0x02, 0xaa, 0x01,
				// Cause5GSM      : &ie.Cause5GSM                  0x59    TV,       2B, 9.11.4.2
				0x59, 0x00,
				// MaxNumOfSupportedPktFilters : &ie.MaxNumOfSupportedPktFilters 0x55    TV,       3B, 9.11.4.9
				// AlwaysonPDUSessReq   : &ie.AlwaysonPDUSessReq   0xB0    TV,       1B, 9.11.4.4
				0xB0,
				// IntegrityProtectionMaxDataRate : &ie.IntegrityProtectionMaxDataRate 0x13    TV,       3B, 9.11.4.7
				0x13, 0xff, 0xff,
				// ReqQosRules          : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				// ReqQosFlowDescs      : &ie.QosFlowDescs         0x79 TLV-E, 6-65538B, 9.11.4.12
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x07,
				0x80,
				0x00, 0x0d, 0x00, // DNS IPv4 Req, 0B
				0x00, 0x10, 0x00, // IPv4 Link MTU Req, 0B
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			input: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
					APMQF:   true,
				},
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: false,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq: true,
						DNSV4Req:       true,
					},
				},
			},
		},
		{
			name: "Case 2 - another 5 optional IEs",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// MaxNumOfSupportedPktFilters : &ie.MaxNumOfSupportedPktFilters 0x55    TV,       3B, 9.11.4.9
				// AlwaysonPDUSessReq   : &ie.AlwaysonPDUSessReq   0xB0    TV,       1B, 9.11.4.4
				// ReqQosRules          : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				0x7A, 0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ReqQosFlowDescs      : &ie.QosFlowDescs         0x79 TLV-E, 6-65538B, 9.11.4.12
				0x79, 0x00, 0x06,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
			},
			input: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
				ReqQosRules: &ie.QosRules{
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
					}},
				},
				ReqQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{{
						QFI:    9,
						OpCode: ie.QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(ie.Qfd_5QI9),
					}},
				},
			},
		},
		{
			name: "Case 3 - another 3 optional IEs",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xc9, // EPD, PDUSessId, PTI, MsgType
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			input: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
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
