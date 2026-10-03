package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNwFeatureSupport5GSMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NwFeatureSupport5GSM
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xf1},
			expected: &NwFeatureSupport5GSM{
				EPTS1: 0x01,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54, 0x32, 0x10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NwFeatureSupport5GSM)
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

func TestNwFeatureSupport5GSMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NwFeatureSupport5GSM
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &NwFeatureSupport5GSM{
				EPTS1: 0x01,
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
