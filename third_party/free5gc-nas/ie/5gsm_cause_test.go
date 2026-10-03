package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCause5GSMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *Cause5GSM
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x64},
			expected: &Cause5GSM{
				Value: Cause5GSM_ConditionalIEErr,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(Cause5GSM)
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

func TestCause5GSMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Cause5GSM
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &Cause5GSM{
				Value: Cause5GSM_PDUSessTypeUnstructuredOnlyAllowed,
			},
			expected: []byte{0x3a},
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
