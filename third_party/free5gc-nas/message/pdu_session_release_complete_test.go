package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessRelCompleteUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x01, 0xbf, 0xd4,
			},
			expected: &PDUSessRelComplete{
				PDUSessId: 1,
				PTI:       191,
			},
		},
		{
			name: "Case full",
			input: []byte{
				0x2e, 0x01, 0xbf, 0xd4,
				// Cause5GSM, TV:
				0x59, 0x24,
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
			},
			expected: &PDUSessRelComplete{
				PDUSessId: 1,
				PTI:       191,
				Cause5GSM: &ie.Cause5GSM{
					Value: ie.Cause5GSM_RegularDeactivation,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessRelComplete)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessRelCompleteMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x01, 0xbf, 0xd4,
			},
			input: &PDUSessRelComplete{
				PDUSessId: 1,
				PTI:       191,
			},
		},
		{
			name: "Case full",
			expected: []byte{
				0x2e, 0x01, 0xbf, 0xd4,
				// Cause5GSM, TV:
				0x59, 0x24,
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x07,
				0x80,
				0x00, 0x0d, 0x00, // DNS IPv4 Req, 0B
				0x00, 0x10, 0x00, // IPv4 Link MTU Req, 0B
			},
			input: &PDUSessRelComplete{
				PDUSessId: 1,
				PTI:       191,
				Cause5GSM: &ie.Cause5GSM{
					Value: ie.Cause5GSM_RegularDeactivation,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq: true,
						DNSV4Req:       true,
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
