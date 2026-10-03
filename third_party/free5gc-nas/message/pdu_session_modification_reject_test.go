package message

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessModRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x05, 0x01, 0xca, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM V
				0x00,
				// BackoffTimerValue : &ie.GPRSTimer3         0x37   TLV,       3B, 9.11.2.5
				0x37, 0x01, 0x06,
				// CongestionReattemptIndicator5GSM : &ie.CongestionReattemptIndicator5GSM 0x61 TLV, 3B, 9.11.4.21
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
				// ReattemptIndicator : &ie.ReattemptIndicator 0x1D   TLV,       3B, 9.11.4.17
			},
			expected: &PDUSessModRej{
				PDUSessId: 5,
				PTI:       1,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				BackoffTimerValue: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
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
			msg := new(PDUSessModRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessModRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x05, 0x01, 0xca, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM V
				0x00,
				// BackoffTimerValue : &ie.GPRSTimer3         0x37   TLV,       3B, 9.11.2.5
				0x37, 0x01, 0x06,
				// CongestionReattemptIndicator5GSM : &ie.CongestionReattemptIndicator5GSM 0x61 TLV, 3B, 9.11.4.21
				// ie.ExtendedProtCfgOpts    TLV-E, 4-65538B, 9.11.4.6
				0x7B, 0x00, 0x0d,
				0x80,
				0x00, 0x0d, 0x04, // DNS IPv4 , 4B
				0x08, 0x08, 0x08, 0x08, // 8.8.8.8
				0x00, 0x10, 0x02, // IPv4 Link MTU, 2B
				0x05, 0x78, // 1400
				// ReattemptIndicator : &ie.ReattemptIndicator 0x1D   TLV,       3B, 9.11.4.17
			},
			input: &PDUSessModRej{
				PDUSessId: 5,
				PTI:       1,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				BackoffTimerValue: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
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
