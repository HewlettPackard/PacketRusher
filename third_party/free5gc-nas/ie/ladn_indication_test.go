package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLADNIndUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *LADNInd
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x04, 0x03, 0x61, 0x62, 0x63},
			expected: &LADNInd{
				DNNs: []DNN{
					{
						Value: "abc",
					},
				},
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x04, 0x03, 0x61, 0x62, 0x63, 0x04, 0x3, 0x63, 0x6f, 0x6d},
			expected: &LADNInd{
				DNNs: []DNN{
					{
						Value: "abc",
					},
					{
						Value: "com",
					},
				},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x02, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
		{
			name:  "Negative Case 2",
			input: []byte{0x32, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(LADNInd)
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

func TestLADNIndMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *LADNInd
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x04, 0x03, 0x61, 0x62, 0x63},
			input: &LADNInd{
				DNNs: []DNN{
					{
						Value: "abc",
					},
				},
			},
		},
		{
			name:     "Positive Case 2",
			expected: []byte{0x04, 0x03, 0x61, 0x62, 0x63, 0x04, 0x3, 0x63, 0x6f, 0x6d},
			input: &LADNInd{
				DNNs: []DNN{
					{
						Value: "abc",
					},
					{
						Value: "com",
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
