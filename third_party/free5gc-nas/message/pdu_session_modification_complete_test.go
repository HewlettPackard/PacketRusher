package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessModCompleteUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x01, 0xbf, 0xcc, // EPD, SecHdr/Spare, Msg Type
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
				// ie.PortMgmtInfoCntr 0x74 TLV-E, 4-65538B, 9.11.4.27
			},
			expected: &PDUSessModComplete{
				PDUSessId: 1,
				PTI:       191,
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq:        true,
						DNSV4Req:              true,
						UEStatus3GPPPSDataOff: ie.UEStatus3GPPPSDataOff_Deactivate,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessModComplete)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessModCompleteMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x01, 0xbf, 0xcc, // EPD, SecHdr/Spare, Msg Type
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
				// ie.PortMgmtInfoCntr 0x74 TLV-E, 4-65538B, 9.11.4.27
			},
			input: &PDUSessModComplete{
				PDUSessId: 1,
				PTI:       191,
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1400,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
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
