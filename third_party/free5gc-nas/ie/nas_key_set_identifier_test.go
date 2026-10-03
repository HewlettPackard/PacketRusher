package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNASKeySetIdUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NASKeySetId
	}{
		{
			name:  "Native type",
			input: []byte{0x01},
			expected: &NASKeySetId{
				Tsc: SecCtxTypeNative,
				Ksi: 1,
			},
		},
		{
			name:  "Mapped type",
			input: []byte{0x0E},
			expected: &NASKeySetId{
				Tsc: SecCtxTypeMapped,
				Ksi: 6,
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NASKeySetId)
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

func TestNASKeySetIdMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NASKeySetId
		expected []byte
	}{
		{
			name: "Native type",
			input: &NASKeySetId{
				Tsc: SecCtxTypeNative,
				Ksi: 1,
			},
			expected: []byte{0x01},
		},
		{
			name: "Mapped type",
			input: &NASKeySetId{
				Tsc: SecCtxTypeMapped,
				Ksi: 6,
			},
			expected: []byte{0x0E},
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
