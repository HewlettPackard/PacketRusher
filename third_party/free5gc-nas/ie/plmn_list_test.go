package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPLMNListUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *PLMNList
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76, 0x54, 0x32, 0x98, 0x76, 0x54},
			expected: &PLMNList{
				PlmnIds: []PlmnId{
					{
						MCC: "674",
						MNC: "235",
					},
					{
						MCC: "896",
						MNC: "457",
					},
				},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PLMNList)
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

func TestPLMNListMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *PLMNList
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &PLMNList{
				PlmnIds: []PlmnId{
					{
						MCC: "674",
						MNC: "235",
					},
					{
						MCC: "896",
						MNC: "457",
					},
				},
			},
			expected: []byte{0x76, 0x54, 0x32, 0x98, 0x76, 0x54},
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
