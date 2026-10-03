package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrackingAreaId5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *TrackingAreaId5GS
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc},
			expected: &TrackingAreaId5GS{
				PlmnId: PlmnId{
					MCC: "674",
					MNC: "235",
				},
				TAC: "10fedc",
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(TrackingAreaId5GS)
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

func TestTrackingAreaId5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *TrackingAreaId5GS
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x76, 0x54, 0x32, 0x10, 0xfe, 0xdc},
			input: &TrackingAreaId5GS{
				PlmnId: PlmnId{
					MCC: "674",
					MNC: "235",
				},
				TAC: "10fedc",
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
