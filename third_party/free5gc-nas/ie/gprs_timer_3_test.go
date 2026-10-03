package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGPRSTimer3UnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *GPRSTimer3
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x76},
			expected: &GPRSTimer3{
				Unit:  TimerIncIn_2Seconds,
				Value: 0x16,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(GPRSTimer3)
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

func TestGPRSTimer3MarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *GPRSTimer3
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &GPRSTimer3{
				Value: 0x16,
				Unit:  TimerIncIn_2Seconds,
			},
			expected: []byte{0x76},
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

func TestGPRSTimer3_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input uint32
		e     *GPRSTimer3
	}{
		{
			name: "Case 2", input: 2,
			e: &GPRSTimer3{Unit: TimerIncIn_2Seconds, Value: 1},
		},
		{
			name: "Case 360", input: 360,
			e: &GPRSTimer3{Unit: TimerIncIn_30Seconds, Value: 12},
		},
		{
			name: "Case 1859", input: 1859,
			e: &GPRSTimer3{Unit: TimerIncIn_1Minute, Value: 30},
		},
		{
			name: "Case 1920", input: 1920,
			e: &GPRSTimer3{Unit: TimerIncIn_10Minutes, Value: 3},
		},
		{
			name: "Case 19200", input: 19200,
			e: &GPRSTimer3{Unit: TimerIncIn_1Hour, Value: 5},
		},
		{
			name: "Case 115200", input: 115200,
			e: &GPRSTimer3{Unit: TimerIncIn_10Hours, Value: 3},
		},
		{
			name: "Case 1151999", input: 1151999,
			e: &GPRSTimer3{Unit: TimerIncIn_10Hours, Value: 31},
		},
		{
			name: "Case 1152000", input: 1152000,
			e: &GPRSTimer3{Unit: TimerIncIn_320Hours, Value: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(GPRSTimer3)
			ie.Set(tc.input)
			require.Equal(t, tc.e, ie)
		})
	}
}
