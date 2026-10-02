package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestPDUSessRelCmdUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x2e, 0x01, 0xbf, 0xd3, // EPD, PDUSessId, PTI, MsgType
				// ie.Cause5GSM                             V,       1B, 9.11.4.2
				0x00,
				// ie.GPRSTimer3                          TLV,       3B, 9.11.2.5
				0x37, 0x01, 0x06,
				// ie.EAPMsg                            TLV-E,  7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
				// ie.CongestionReattemptIndicator5GSM    TLV,       3B, 9.11.4.21
				// ie.ExtendedProtCfgOpts                TLV-E, 4-65538B, 9.11.4.6
				// ie.AccessType                           TV,       1B, 9.11.2.1A
			},
			expected: &PDUSessRelCmd{
				PDUSessId: 1,
				PTI:       191,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
				},
				BackoffTimerValue: &ie.GPRSTimer3{
					Unit:  ie.TimerIncIn_10Minutes,
					Value: 6,
				},
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{
						0x01, 0x02, 0x03, 0x04,
					},
				},
				// CongestionReattemptIndicator5GSM: &ie.CongestionReattemptIndicator5GSM{},
				// ExtendedProtCfgOpts              : &ie.ExtendedProtCfgOpts{},
				// AccessType                      : &ie.AccessType{},

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(PDUSessRelCmd)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestPDUSessRelCmdMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x2e, 0x01, 0xbf, 0xd3, 0x00,
			},
			input: &PDUSessRelCmd{
				PDUSessId: 1,
				PTI:       191,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x0,
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
