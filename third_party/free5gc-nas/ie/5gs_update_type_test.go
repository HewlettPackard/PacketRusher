package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateType5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *UpdateType5GS
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &UpdateType5GS{
				EPS_PNB_CIoT: 0x03,
				PNB_CIoT_5GS: 0x01,
				NG_RAN_RCU:   true,
				SMSRequested: false,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(UpdateType5GS)
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

func TestUpdateType5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *UpdateType5GS
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &UpdateType5GS{
				EPS_PNB_CIoT: 0x03,
				PNB_CIoT_5GS: 0x01,
				NG_RAN_RCU:   true,
				SMSRequested: false,
			},
			expected: []byte{0x36},
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
