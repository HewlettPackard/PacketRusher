package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestAuthRejUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "5GS AuthRej",
			input: []byte{
				0x7e, 0x00, 0x58, // EPD, SecHdr/Spare, Msg Type
				// EAPMsg : &ie.EAPMsg 0x78 TLV-E, 7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			expected: &AuthRej{
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{0x01, 0x02, 0x03, 0x04},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(AuthRej)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestAuthRejMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "5GS AuthRej",
			expected: []byte{
				0x7e, 0x00, 0x58, // EPD, SecHdr/Spare, Msg Type
				// EAPMsg : &ie.EAPMsg 0x78 TLV-E, 7-1503B, 9.11.2.2
				0x78, 0x00, 0x04,
				0x01, 0x02, 0x03, 0x04,
			},
			input: &AuthRej{
				EAPMsg: &ie.EAPMsg{
					Eap: []byte{0x01, 0x02, 0x03, 0x04},
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
