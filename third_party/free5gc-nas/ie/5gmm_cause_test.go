package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCause5GMMUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *Cause5GMM
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x64},
			expected: &Cause5GMM{
				Value: Cause5GMM_ConditionalIEErr,
			},
		},
		{
			name:  "All other values should be treated as protocol error, unspecified",
			input: []byte{0x71},
			expected: &Cause5GMM{
				Value: Cause5GMM_ProtError,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(Cause5GMM)
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

func TestCause5GMMMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Cause5GMM
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &Cause5GMM{
				Value: Cause5GMM_PayloadWasNotForwarded,
			},
			expected: []byte{0x5a},
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

func TestCause5GMMString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *Cause5GMM
		expected string
	}{
		{
			name: "Positive Case 1",
			input: &Cause5GMM{
				Value: Cause5GMM_PayloadWasNotForwarded,
			},
			expected: "Payload Was Not Forwarded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.input.String()
			require.Equal(t, tc.expected, out)
		})
	}
}
