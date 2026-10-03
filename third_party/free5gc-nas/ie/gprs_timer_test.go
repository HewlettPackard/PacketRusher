package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGPRSTimerUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *GPRSTimer
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x56},
			expected: &GPRSTimer{
				Unit:       TimerIncIn_Decihours,
				TimerValue: 0x16,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(GPRSTimer)
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

func TestGPRSTimerMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *GPRSTimer
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &GPRSTimer{
				TimerValue: 0x16,
				Unit:       TimerIncIn_Decihours,
			},
			expected: []byte{0x56},
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
