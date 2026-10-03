package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGPRSTimer2UnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *GPRSTimer2
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x56},
			expected: &GPRSTimer2{
				Unit:  TimerIncIn_Decihours,
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
			ie := new(GPRSTimer2)
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

func TestGPRSTimer2MarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *GPRSTimer2
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &GPRSTimer2{
				Value: 0x16,
				Unit:  TimerIncIn_Decihours,
			},
			expected: []byte{0x56},
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

func TestGPRSTimer_GprsTmr1ToSec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    uint8
		expected uint32
	}{
		{name: "case 0", input: 0, expected: 0},
		{name: "case 1", input: 1, expected: 2},
		{name: "case 2", input: 2, expected: 4},
		{name: "case 0xf", input: 0xf, expected: 30},
		{name: "case 0x1f", input: 0x1f, expected: 62},
		{name: "case 0x21", input: 0x21, expected: 60},
		{name: "case 0x22", input: 0x22, expected: 120},
		{name: "case 0x23", input: 0x23, expected: 180},
		{name: "case 0x3f", input: 0x3f, expected: 1860},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := GPRSTmr1ToSec(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestGPRSTimerSecToGPRSTmr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    uint32
		expected uint8
	}{
		{name: "case 0", input: 0, expected: 0xe0},
		{name: "case 1", input: 1, expected: 0},
		{name: "case 2", input: 2, expected: 1},
		{name: "case 3", input: 3, expected: 1},
		{name: "case 4", input: 4, expected: 2},
		{name: "case 5", input: 5, expected: 2},
		{name: "case 6", input: 6, expected: 3},
		{name: "case 7", input: 7, expected: 3},
		{name: "case 8", input: 8, expected: 4},
		{name: "case 9", input: 9, expected: 4},
		{name: "case 10", input: 10, expected: 5},
		{name: "case 31", input: 31, expected: 0xf},
		{name: "case 32", input: 32, expected: 0x10},
		{name: "case 61", input: 61, expected: 0x1e},
		{name: "case 62", input: 62, expected: 0x1f},
		{name: "case 63", input: 63, expected: 0x1f},
		{name: "case 64", input: 64, expected: 0x21},
		{name: "case 65", input: 65, expected: 0x21},
		{name: "case 66", input: 66, expected: 0x21},
		{name: "case 119", input: 119, expected: 0x21},
		{name: "case 120", input: 120, expected: 0x22},
		{name: "case 359", input: 359, expected: 0x25},
		{name: "case 360", input: 360, expected: 0x26},
		{name: "case 1859", input: 1859, expected: 0x3e},
		{name: "case 1860", input: 1860, expected: 0x3f},
		{name: "case 1919", input: 1919, expected: 0x3f},
		{name: "case 1920", input: 1920, expected: 0x45},
		{name: "case 11158", input: 11158, expected: 0x5e},
		{name: "case 11159", input: 11159, expected: 0x5e},
		{name: "case 11160", input: 11160, expected: 0x5f},
		{name: "case 11161", input: 11161, expected: 0x5f},
		{name: "case 222222", input: 222222, expected: 0x5f},
		{name: "case 333333", input: 333333, expected: 0x5f},
		{name: "case 444444", input: 444444, expected: 0x5f},
		{name: "case 555555", input: 555555, expected: 0x5f},
		{name: "case 666666", input: 666666, expected: 0x5f},
		{name: "case 777777", input: 777777, expected: 0x5f},
		{name: "case 888888", input: 888888, expected: 0x5f},
		{name: "case 999999", input: 999999, expected: 0x5f},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := SecToGPRSTmr1(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestGPRSTimer2_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input uint32
		e     *GPRSTimer2
	}{
		{
			name: "Case 2", input: 2,
			e: &GPRSTimer2{Unit: TimerIncIn_2Secs, Value: 1},
		},
		{
			name: "Case 360", input: 360,
			e: &GPRSTimer2{Unit: TimerIncIn_1Min, Value: 6},
		},
		{
			name: "Case 1859", input: 1859,
			e: &GPRSTimer2{Unit: TimerIncIn_1Min, Value: 0x1e},
		},
		{
			name: "Case 1920", input: 1920,
			e: &GPRSTimer2{Unit: TimerIncIn_Decihours, Value: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(GPRSTimer2)
			ie.Set(tc.input)
			require.Equal(t, tc.e, ie)
		})
	}
}
