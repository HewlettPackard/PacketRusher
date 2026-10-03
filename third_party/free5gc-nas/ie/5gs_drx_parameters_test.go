package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDRXParams5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *DRXParams5GS
	}{
		{
			name:  "Positive Case",
			input: []byte{0x31},
			expected: &DRXParams5GS{
				Value: DRXCycleParameterT32,
			},
		},
		{
			name:  "Treat value out of range as unspecified",
			input: []byte{0x76},
			expected: &DRXParams5GS{
				Value: DRXValueNotSpecified,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(DRXParams5GS)
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

func TestDRXParams5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *DRXParams5GS
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &DRXParams5GS{
				Value: 0x16,
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

func TestDRXValueString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    DRXValue
		expected string
	}{
		{
			name:     "DRXCycleParameterT32",
			input:    DRXCycleParameterT32,
			expected: "T=32",
		},
		{
			name:     "DRXCycleParameterT64",
			input:    DRXCycleParameterT64,
			expected: "T=64",
		},
		{
			name:     "DRXCycleParameterT128",
			input:    DRXCycleParameterT128,
			expected: "T=128",
		},
		{
			name:     "DRXCycleParameterT256",
			input:    DRXCycleParameterT256,
			expected: "T=256",
		},
		{
			name:     "DRXValueNotSpecified",
			input:    DRXValueNotSpecified,
			expected: "NotSpecified",
		},
		{
			name:     "DRXCycleParameterMAX",
			input:    DRXCycleParameterMAX,
			expected: "T=256",
		},
		{
			name:     "Unknown",
			input:    0x05,
			expected: "Unknown(5)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.input.String())
		})
	}
}
