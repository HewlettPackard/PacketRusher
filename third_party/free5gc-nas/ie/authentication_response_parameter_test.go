package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthRspParamUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *AuthRspParam
	}{
		{
			name: "Data length 16",
			input: []byte{
				0x0B, 0x32, 0x64, 0x66, 0xFB, 0xC0, 0x80, 0x00,
				0x84, 0x9E, 0x28, 0x49, 0x5F, 0xB1, 0x71, 0xA1,
			},
			expected: &AuthRspParam{
				Res: []byte{
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
			ie := new(AuthRspParam)
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

func TestAuthRspParamMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *AuthRspParam
		expected []byte
	}{
		{
			name: "Data length 16",
			input: &AuthRspParam{
				Res: []byte{
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
			input: &AuthRspParam{
				Res: []byte{0x01},
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
