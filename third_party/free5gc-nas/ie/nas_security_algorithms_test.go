package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNASSecAlgosUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *NASSecAlgos
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x43},
			expected: &NASSecAlgos{
				CipheringAlgo: EncAlgo_5GEA4,
				MsgIntAlgo:    IntegrityAlgo_1285GIA3,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x02, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
		{
			name:  "Negative Case 2",
			input: []byte{0x32, 0x01},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(NASSecAlgos)
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

func TestNASSecAlgosMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *NASSecAlgos
		expected []byte
	}{
		{
			name:     "Positive Case 1",
			expected: []byte{0x72},
			input: &NASSecAlgos{
				CipheringAlgo: EncAlgo_5GEA7,
				MsgIntAlgo:    IntegrityAlgo_1285GIA2,
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
