package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessModCmdUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM TV
				0x59, 0x00,
				// ie.SessAMBR TLV
				0x2A, 0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// AuthoQosRules        : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				0x7A, 0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.RQTimerValue TV
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
			},
			expected: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
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
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1400,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
			},
		},
		{
			name: "Case 2 another 5 optional IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// RQTimerValue         : &ie.GPRSTimer            0x56    TV,       2B, 9.11.2.3
				0x56, 0x56,
				// AlwaysonPDUSessInd   : &ie.AlwaysonPDUSessInd   0x80    TV,       1B, 9.11.4.3
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
				// ATSSSCntr            : &ie.ATSSSCntr            0x77 TLV-E, 3-65538B, 9.11.4.22
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
			},
			expected: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
				RQTimerValue: &ie.GPRSTimer{
					Unit:       ie.TimerIncIn_Decihours,
					TimerValue: 0x16,
				},
			},
		},
		{
			name: "Case 3 another 3 optional IEs",
			input: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// ServingPLMNRateCtrl  : &ie.ServingPLMNRateCtrl  0x1E   TLV,       4B, 9.11.4.20
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			expected: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessModCmd)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessModCmdMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM TV
				0x59, 0x00,
				// ie.SessAMBR TLV
				0x2A, 0x06,
				0x06, 0x00, 0x64, 0x06, 0x00, 0xc8,
				// AuthoQosRules        : &ie.QosRules             0x7A TLV-E, 7-65538B, 9.11.4.13
				0x7A, 0x00, 0x09,
				0x01, 0x00, 0x06, 0x31, 0x31, 0x01, 0x01, 0xff, 0x09,
				// ie.RQTimerValue TV
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
			},
			input: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
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
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1400,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
			},
		},
		{
			name: "Case 2 another 5 optional IEs",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// RQTimerValue         : &ie.GPRSTimer            0x56    TV,       2B, 9.11.2.3
				0x56, 0x56,
				// AlwaysonPDUSessInd   : &ie.AlwaysonPDUSessInd   0x80    TV,       1B, 9.11.4.3
				// MappedEPSBearerCtxs   : &ie.MappedEPSBearerCtxs   0x75 TLV-E, 7-65538B, 9.11.4.8
				// ATSSSCntr            : &ie.ATSSSCntr            0x77 TLV-E, 3-65538B, 9.11.4.22
				// IPHdrCompressionCfg  : &ie.IPHdrCompressionCfg  0x66   TLV,   5-257B, 9.11.4.24
			},
			input: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
				RQTimerValue: &ie.GPRSTimer{
					Unit:       ie.TimerIncIn_Decihours,
					TimerValue: 0x16,
				},
			},
		},
		{
			name: "Case 3 another 3 optional IEs",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xcb, // EPD, PDUSessId, PTI, MsgType
				// PortMgmtInfoCntr     : &ie.PortMgmtInfoCntr     0x74 TLV-E, 4-65538B, 9.11.4.27
				// ServingPLMNRateCtrl  : &ie.ServingPLMNRateCtrl  0x1E   TLV,       4B, 9.11.4.20
				// EthHdrCompressionCfg : &ie.EthHdrCompressionCfg 0x1F   TLV,       3B, 9.11.4.28
			},
			input: &PDUSessModCmd{
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
