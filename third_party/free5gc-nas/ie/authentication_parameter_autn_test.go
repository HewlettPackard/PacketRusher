package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthParamAUTNUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *AuthParamAUTN
	}{
		{
			name: "Data length 16",
			input: []byte{
				0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
				0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1, 0x71, 0xA1,
			},
			expected: &AuthParamAUTN{
				Autn: []byte{
					0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
					0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1, 0x71, 0xA1,
				},
			},
		},
		{
			name:     "Data length 1",
			input:    []byte{0x01},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(AuthParamAUTN)
			err := ie.UnmarshalBinary(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAuthParamAUTNMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *AuthParamAUTN
		expected []byte
	}{
		{
			name: "Data length 16",
			input: &AuthParamAUTN{
				Autn: []byte{
					0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
					0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1, 0x71, 0xA1,
				},
			},
			expected: []byte{
				0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
				0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1, 0x71, 0xA1,
			},
		},
		{
			name: "Data length 1",
			input: &AuthParamAUTN{
				Autn: []byte{0x01},
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}
