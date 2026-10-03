package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeregTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *DeregType
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &DeregType{
				Switchoff:     false,
				ReregRequired: true,
				AccessType:    2,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(DeregType)
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

func TestDeregTypeMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *DeregType
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &DeregType{
				Switchoff:     false,
				ReregRequired: true,
				AccessType:    2,
			},
			expected: []byte{0x06},
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
