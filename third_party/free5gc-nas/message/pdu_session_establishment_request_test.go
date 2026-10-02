package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessEstReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xc1, // EPD, PDUSessId, PTI, MsgType
				// ie.IntegrityProtectionMaxDataRate V
				0xff, 0xff,
				// ie.PDUSessType              TV, 9-
				0x91,
				// ie.SSCMode                  TV, A-
				0xa1,
				// ie.Capability5GSM          TLV,    3-15B, 9.11.4.1
				0x28, 0x01, 0xaa,
				// ie.MaxNumOfSupportedPktFilters     TV,       3B, 9.11.4.9
				// 0x55
				// ie.AlwaysonPDUSessReq       TV,       1B, 9.11.4.4
				0xB1,
				// ie.SMPDUDNReqCntr          TLV,   3-255B, 9.11.4.15
				// 0x39
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
				// ie.IPHdrCompressionCfg     TLV,   5-257B, 9.11.4.24
				// 0x66
				// ie.DSTTEthPortMACAddr      TLV,       8B, 9.11.4.25
				// 0x6E
				// ie.UEDSTTResidenceTime     TLV,      10B, 9.11.4.26
				// 0x6F
				// ie.PortMgmtInfoCntr      TLV-E, 8-65538B, 9.11.4.27
				// 0x74
				// ie.EthHdrCompressionCfg    TLV,       3B, 9.11.4.28
				// 0x1F
				// ie.PDUAddr                 TLV,      11B, 9.11.4.10
				0x29, 0x1d,
				0x0b,
				0xaa, 0xbb, 0xcc, 0xdd, 0x12, 0x34, 0x56, 0x78,
				0xde, 0xad, 0xbe, 0xef,
				0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
				0x78, 0x56, 0x34, 0x12, 0x12, 0x34, 0x56, 0x78,
			},
			expected: &PDUSessEstReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				PDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				SSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: true,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq:        true,
						DNSV4Req:              true,
						UEStatus3GPPPSDataOff: ie.UEStatus3GPPPSDataOff_Deactivate,
					},
				},
				SuggestedIfId: &ie.PDUAddr{
					IPv6IfId: []byte{
						0xaa, 0xbb, 0xcc, 0xdd, 0x12, 0x34, 0x56, 0x78,
					},
					IPv4: []byte{0xde, 0xad, 0xbe, 0xef},
					SMFIPv6LLA: []byte{
						0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
						0x78, 0x56, 0x34, 0x12, 0x12, 0x34, 0x56, 0x78,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessEstReq)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessEstReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xc1, // EPD, PDUSessId, PTI, MsgType
				// ie.IntegrityProtectionMaxDataRate V
				0xff, 0xff,
				// ie.PDUSessType              TV, 9-
				0x91,
				// ie.SSCMode                  TV, A-
				0xa1,
				// ie.Capability5GSM          TLV,    3-15B, 9.11.4.1
				0x28, 0x02, 0xaa, 0x00,
				// ie.MaxNumOfSupportedPktFilters     TV,       3B, 9.11.4.9
				// 0x55
				// ie.AlwaysonPDUSessReq       TV,       1B, 9.11.4.4
				0xB1,
				// ie.SMPDUDNReqCntr          TLV,   3-255B, 9.11.4.15
				// 0x39
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4, 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
				// ie.IPHdrCompressionCfg     TLV,   5-257B, 9.11.4.24
				// 0x66
				// ie.DSTTEthPortMACAddr      TLV,       8B, 9.11.4.25
				// 0x6E
				// ie.UEDSTTResidenceTime     TLV,      10B, 9.11.4.26
				// 0x6F
				// ie.PortMgmtInfoCntr      TLV-E, 8-65538B, 9.11.4.27
				// 0x74
				// ie.EthHdrCompressionCfg    TLV,       3B, 9.11.4.28
				// 0x1F
				// ie.PDUAddr                 TLV,      11B, 9.11.4.10
				0x29, 0x1d,
				0x0b,
				0xaa, 0xbb, 0xcc, 0xdd, 0x12, 0x34, 0x56, 0x78,
				0xde, 0xad, 0xbe, 0xef,
				0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
				0x78, 0x56, 0x34, 0x12, 0x12, 0x34, 0x56, 0x78,
			},
			input: &PDUSessEstReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				PDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				SSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: true,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1400,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
				SuggestedIfId: &ie.PDUAddr{
					IPv6IfId: []byte{
						0xaa, 0xbb, 0xcc, 0xdd, 0x12, 0x34, 0x56, 0x78,
					},
					IPv4: []byte{0xde, 0xad, 0xbe, 0xef},
					SMFIPv6LLA: []byte{
						0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
						0x78, 0x56, 0x34, 0x12, 0x12, 0x34, 0x56, 0x78,
					},
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
