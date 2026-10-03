package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestAuthFailureUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected Message
	}{
		{
			name: "Case 1",
			input: []byte{
				0x7e, 0x00, 0x59, 0x47,
			},
			expected: &AuthFailure{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
			},
		},
		{
			name: "Case 2",
			input: []byte{
				0x7e, 0x00, 0x59, // EPD, SecHdr/Spare, Msg Type
				// ie.Cause5GMM V
				0x47,
				// AuthFailureParam TLV
				0x30, 0x0E,
				0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
				0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1,
			},
			expected: &AuthFailure{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				AuthFailureParam: &ie.AuthFailureParam{
					Value: []byte{
						0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
						0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(AuthFailure)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestAuthFailureMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    Message
		expected []byte
	}{
		{
			name: "Case 1",
			expected: []byte{
				0x7e, 0x00, 0x59, 0x47,
			},
			input: &AuthFailure{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
			},
		},
		{
			name: "Case 2",
			expected: []byte{
				0x7e, 0x00, 0x59, 0x47,
				// AuthFailureParam:
				0x30, 0x0E, // IEI, Length
				0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
				0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1,
			},
			input: &AuthFailure{
				Cause5GMM: &ie.Cause5GMM{
					Value: ie.Cause5GMM_NgksiAlreadyInUse,
				},
				AuthFailureParam: &ie.AuthFailureParam{
					Value: []byte{
						0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
						0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1,
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
