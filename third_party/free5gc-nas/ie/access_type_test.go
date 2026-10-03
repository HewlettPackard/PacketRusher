package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccessTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *AccessType
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x00},
			expected: &AccessType{
				Value: 0x00,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x01},
			expected: &AccessType{
				Value: 0x01,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(AccessType)
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

func TestAccessTypeMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *AccessType
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &AccessType{
				Value: 0x00,
			},
			expected: []byte{0x00},
		},
		{
			name: "Positive Case 2",
			input: &AccessType{
				Value: 0x01,
			},
			expected: []byte{0x01},
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
