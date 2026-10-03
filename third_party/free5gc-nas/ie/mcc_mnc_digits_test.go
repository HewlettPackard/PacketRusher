package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlmnIdDigitsUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PlmnId
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76, 0x54, 0x32},
			expected: &PlmnId{
				MCC: "674",
				MNC: "235",
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PlmnId)
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

func TestPlmnIdDigitsMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PlmnId
		expected []byte
	}{
		{
			name: "Positive Case - MNC 3 digits 674235",
			input: &PlmnId{
				MCC: "674",
				MNC: "235",
			},
			expected: []byte{0x76, 0x54, 0x32},
		},
		{
			name: "Positive Case - MNC 3 digits 999002",
			input: &PlmnId{
				MCC: "999",
				MNC: "002",
			},
			expected: []byte{0x99, 0x29, 0x00},
		},
		{
			name: "Positive Case - MNC 2 digits",
			input: &PlmnId{
				MCC: "674",
				MNC: "23",
			},
			expected: []byte{0x76, 0xf4, 0x32},
		},
		// Testing devices could have MCC < 100
		// {
		// 	name: "Invalid Case 1.1 - MCC should be 3 digits",
		// 	input: &PlmnId{
		// 		MCC: "67",
		// 		MNC: "235",
		// 	},
		// },
		// {
		// 	name: "Invalid Case 1.2 - MCC should be 3 digits",
		// 	input: &PlmnId{
		// 		MCC: "6",
		// 		MNC: "235",
		// 	},
		// },
		// {
		// 	name: "Invalid Case 1.3 - MCC should be 3 digits",
		// 	input: &PlmnId{
		// 		MCC: "6740",
		// 		MNC: "235",
		// 	},
		// },
		{
			name: "Invalid Case 2.1 - MNC should be 2 or 3 digits",
			input: &PlmnId{
				MCC: "674",
				MNC: "2",
			},
		},
		{
			name: "Invalid Case 2.2 - MNC should be 2 or 3 digits",
			input: &PlmnId{
				MCC: "674",
				MNC: "0350",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := []byte{0, 0, 0}
			err := tc.input.MarshalBinary(b)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}
