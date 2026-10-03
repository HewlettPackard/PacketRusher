package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessRelRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xd2, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM V
				0x00,
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
			},
			expected: &PDUSessRelRej{
				PDUSessId: 5,
				PTI:       1,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessRelRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessRelRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xd2, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM V
				0x00,
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
			},
			input: &PDUSessRelRej{
				PDUSessId: 5,
				PTI:       1,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}
