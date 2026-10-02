package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestStatus5GSMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "5GSM Status",
			input: []byte{
				0x2e, 0x01, 0x01, 0xd6, // EPD, PDUSessID, PTI, Msg Type
				0x32, // Cause5GSM (PDUSessTypeIpv4OnlyAllowed)
			},
			expected: &Status5GSM{
				PDUSessId: 0x01,
				PTI:       0x01,
				Cause5GSM: &ie.Cause5GSM{
					Value: ie.Cause5GSM_PDUSessTypeIpv4OnlyAllowed,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(Status5GSM)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestStatus5GSMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "5GSM Status",
			expected: []byte{
				0x2e, 0x01, 0x01, 0xd6, // EPD, PDUSessID, PTI, Msg Type
				0x32, // Cause5GSM (PDUSessTypeIpv4OnlyAllowed)
			},
			input: &Status5GSM{
				PDUSessId: 0x01,
				PTI:       0x01,
				Cause5GSM: &ie.Cause5GSM{
					Value: ie.Cause5GSM_PDUSessTypeIpv4OnlyAllowed,
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
