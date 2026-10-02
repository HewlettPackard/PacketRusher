package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlwaysonPDUSessReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *AlwaysonPDUSessReq
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0xb0},
			expected: &AlwaysonPDUSessReq{
				APSR: false,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0xb1},
			expected: &AlwaysonPDUSessReq{
				APSR: true,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(AlwaysonPDUSessReq)
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

func TestAlwaysonPDUSessReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *AlwaysonPDUSessReq
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &AlwaysonPDUSessReq{
				APSR: false,
			},
			expected: []byte{0x00},
		},
		{
			name: "Positive Case 2",
			input: &AlwaysonPDUSessReq{
				APSR: true,
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
