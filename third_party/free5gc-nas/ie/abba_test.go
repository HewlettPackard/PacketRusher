package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestABBAUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *ABBA
	}{
		{
			name:  "Data length 2",
			input: []byte{0x01, 0x02},
			expected: &ABBA{
				Abba: []byte{0x01, 0x02},
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
			ie := new(ABBA)
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

func TestABBAMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *ABBA
		expected []byte
	}{
		{
			name: "Data length 2",
			input: &ABBA{
				Abba: []byte{0x01, 0x02},
			},
			expected: []byte{0x01, 0x02},
		},
		{
			name: "Data length 1",
			input: &ABBA{
				Abba: []byte{0x01},
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
