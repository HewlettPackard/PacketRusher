package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestSecModeRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x5f, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
			},
			expected: &SecModeRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(SecModeRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestSecModeRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x5f, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
			},
			input: &SecModeRej{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
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
