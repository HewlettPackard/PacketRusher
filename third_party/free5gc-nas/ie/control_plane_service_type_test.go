package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCtrlPlaneSvcTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *CtrlPlaneSvcType
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x71},
			expected: &CtrlPlaneSvcType{
				Value: 1,
			},
		},
		{
			name:  "Positive Case 2",
			input: []byte{0x66},
			expected: &CtrlPlaneSvcType{
				Value: 6,
			},
		},
		{
			name:  "Positive Case 3",
			input: []byte{0x6d},
			expected: &CtrlPlaneSvcType{
				Value: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(CtrlPlaneSvcType)
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

func TestCtrlPlaneSvcTypeMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *CtrlPlaneSvcType
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &CtrlPlaneSvcType{
				Value: 1,
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
