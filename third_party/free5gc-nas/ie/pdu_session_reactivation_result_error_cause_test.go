package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPDUSessReactivationResultErrCauseUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PDUSessReactivationResultErrCause
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76, 0x54},
			expected: &PDUSessReactivationResultErrCause{
				IdCause: []SessIdCausePair{
					{
						PDUSessID: 0x76,
						Value:     0x54,
					},
				},
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x76, 0x54, 0x32, 0x10},
			expected: &PDUSessReactivationResultErrCause{
				IdCause: []SessIdCausePair{
					{
						PDUSessID: 0x76,
						Value:     0x54,
					},
					{
						PDUSessID: 0x32,
						Value:     0x10,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PDUSessReactivationResultErrCause)
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

func TestPDUSessReactivationResultErrCauseMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PDUSessReactivationResultErrCause
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x76, 0x54},
			input: &PDUSessReactivationResultErrCause{
				IdCause: []SessIdCausePair{
					{
						PDUSessID: 0x76,
						Value:     0x54,
					},
				},
			},
		},
		{
			name:     "Positive Case 2",
			expected: []byte{0x76, 0x54, 0x32, 0x10},
			input: &PDUSessReactivationResultErrCause{
				IdCause: []SessIdCausePair{
					{
						PDUSessID: 0x76,
						Value:     0x54,
					},
					{
						PDUSessID: 0x32,
						Value:     0x10,
					},
				},
			},
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
