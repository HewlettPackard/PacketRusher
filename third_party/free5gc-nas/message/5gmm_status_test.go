package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestStatus5GMMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "5GMM Status",
			input: []byte{
				0x7e, 0x00, 0x65, // EPD, SecHdr/Spare, Msg Type
				0x03, // Cause5GMM (Illegal UE)
			},
			expected: &Status5GMM{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_IllegalUE,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(Status5GMM)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestStatus5GMMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "5GMM Status",
			expected: []byte{
				0x7e, 0x00, 0x64, // EPD, SecHdr/Spare, Msg Type
				0x03, // Cause5GMM (Illegal UE)
			},
			input: &Status5GMM{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_IllegalUE,
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
