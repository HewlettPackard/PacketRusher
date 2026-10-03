package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIMEISVReqUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *IMEISVReq
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x71},
			expected: &IMEISVReq{
				Value: IMEISV_Requested,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(IMEISVReq)
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

func TestIMEISVReqMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *IMEISVReq
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &IMEISVReq{
				Value: IMEISV_Requested,
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
