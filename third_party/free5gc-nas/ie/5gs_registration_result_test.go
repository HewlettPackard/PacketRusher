package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegResult5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *RegResult5GS
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &RegResult5GS{
				DisasterRoamingRegResult: true,
				EmergRegistered:          true,
				NSSAAPerformed:           true,
				SMSAllowed:               false,
				Value:                    0x06,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(RegResult5GS)
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

func TestRegResult5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *RegResult5GS
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &RegResult5GS{
				DisasterRoamingRegResult: false,
				EmergRegistered:          true,
				NSSAAPerformed:           true,
				SMSAllowed:               false,
				Value:                    0x06,
			},
			expected: []byte{0x36},
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
