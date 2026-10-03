package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthParamRANDUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *AuthParamRAND
	}{
		{
			name: "Data length 16",
			input: []byte{
				0x34, 0xFB, 0x88, 0xAF, 0x54, 0xA9, 0x1A, 0x58,
				0x29, 0xCD, 0x90, 0xC0, 0xA8, 0xFA, 0xD4, 0xDA,
			},
			expected: &AuthParamRAND{
				Rand: []byte{
					0x34, 0xFB, 0x88, 0xAF, 0x54, 0xA9, 0x1A, 0x58,
					0x29, 0xCD, 0x90, 0xC0, 0xA8, 0xFA, 0xD4, 0xDA,
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
			ie := new(AuthParamRAND)
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

func TestAuthParamRANDMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *AuthParamRAND
		expected []byte
	}{
		{
			name: "Data length 16",
			input: &AuthParamRAND{
				Rand: []byte{
					0x34, 0xFB, 0x88, 0xAF, 0x54, 0xA9, 0x1A, 0x58,
					0x29, 0xCD, 0x90, 0xC0, 0xA8, 0xFA, 0xD4, 0xDA,
				},
			},
			expected: []byte{
				0x34, 0xFB, 0x88, 0xAF, 0x54, 0xA9, 0x1A, 0x58,
				0x29, 0xCD, 0x90, 0xC0, 0xA8, 0xFA, 0xD4, 0xDA,
			},
		},
		{
			name: "Data length 1",
			input: &AuthParamRAND{
				Rand: []byte{0x01},
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
