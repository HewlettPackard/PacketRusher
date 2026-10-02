package ie

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestTimeZoneUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *TimeZone
	}{
		{
			name: "+08:30",
			input: []byte{
				0x43,
			},
			expected: &TimeZone{
				Value: "+08:30",
			},
		},
		{
			name: "-04:45",
			input: []byte{
				0x99,
			},
			expected: &TimeZone{
				Value: "-04:45",
			},
		},
		{
			name: "+02:00",
			input: []byte{
				0x80,
			},
			expected: &TimeZone{
				Value: "+02:00",
			},
		},
		{
			name: "-04:00",
			input: []byte{
				0x69,
			},
			expected: &TimeZone{
				Value: "-04:00",
			},
		},
		// Unmarshal has no DST test cases, since DST info can't be retrieved from encoded value
		// e.g. DST(+1) + EST(UTC-5) = EDT(UTC-4)
		// e.g. DST(+1) + CET(UTC+1) = CEST(UTC+2)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(TimeZone)
			err := msg.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, msg)
		})
	}
}

func TestTimeZoneMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *TimeZone
		expected []byte
	}{
		{
			name: "+08:30",
			input: &TimeZone{
				Value: "+08:30",
			},
			expected: []byte{
				0x43,
			},
		},
		{
			name: "-04:45",
			input: &TimeZone{
				Value: "-04:45",
			},
			expected: []byte{
				0x99,
			},
		},
		// with Daylight Saving Time
		{
			name: "+08:30+1",
			input: &TimeZone{
				Value: "+08:30+1",
			},
			expected: []byte{
				0x83,
			},
		},
		{
			name: "-04:45+2",
			input: &TimeZone{
				Value: "-04:45+2",
			},
			expected: []byte{
				0x19,
			},
		},
		{
			name: "-00:30+1",
			input: &TimeZone{
				Value: "-00:30+1",
			},
			expected: []byte{
				0x20,
			},
		},
		{
			name: "+02:00",
			input: &TimeZone{
				Value: "+02:00",
			},
			expected: []byte{
				0x80,
			},
		},
		{
			name: "-04:00",
			input: &TimeZone{
				Value: "-04:00",
			},
			expected: []byte{
				0x69,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, b)
		})
	}
}

func TestTimeZoneOffset(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *TimeZone
		expected int
	}{
		{
			name: "+08:30",
			input: &TimeZone{
				Value: "+08:30",
			},
			expected: 8*3600 + 30*60,
		},
		{
			name: "-04:45",
			input: &TimeZone{
				Value: "-04:45",
			},
			expected: (-1) * (4*3600 + 45*60),
		},
		// with Daylight Saving Time
		{
			name: "+08:30+1",
			input: &TimeZone{
				Value: "+08:30+1",
			},
			expected: 9*3600 + 30*60,
		},
		{
			name: "-04:45+2",
			input: &TimeZone{
				Value: "-04:45+2",
			},
			expected: (-1) * (2*3600 + 45*60),
		},
		{
			name: "-00:30+1",
			input: &TimeZone{
				Value: "-00:30+1",
			},
			expected: (-1) * (-1*3600 + 30*60),
		},
		{
			name: "+02:00",
			input: &TimeZone{
				Value: "+02:00",
			},
			expected: 2 * 3600,
		},
		{
			name: "-04:00",
			input: &TimeZone{
				Value: "-04:00",
			},
			expected: (-1) * (4 * 3600),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			offset, err := tc.input.Offset()
			require.NoError(t, err)
			require.Equal(t, tc.expected, offset)
		})
	}
}

func TestNegativeTimeZoneUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       []byte
		expectedErr error
	}{
		{
			name:        "len is not 1 byte",
			input:       nil,
			expectedErr: errors.Errorf("TimeZone bad IE len(0)"),
		},
		{
			name:        "len is not 1 byte",
			input:       []byte{0x00, 0x00},
			expectedErr: errors.Errorf("TimeZone bad IE len(2)"),
		},
		{
			name:        "decoded result is not a valid timezone: +13",
			input:       []byte{0x25},
			expectedErr: errors.Errorf("TimeZone UnmarshalBinary(): invalid value[13] for hr"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := new(TimeZone)
			err := msg.UnmarshalBinary(tc.input)
			require.EqualError(t, err, tc.expectedErr.Error())
		})
	}
}

func TestNegativeTimeZoneMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *TimeZone
		expected error
	}{
		{
			name: "missing [+-]: 08:30",
			input: &TimeZone{
				Value: "08:30",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[08:30]"),
		},
		{
			name: " missing [+-]: *08:30",
			input: &TimeZone{
				Value: "*08:30",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[*08:30]"),
		},
		{
			name: " multiple [+-]: ++08:30",
			input: &TimeZone{
				Value: "++08:30",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[++08:30]"),
		},
		{
			name: " multiple [+-]: --08:30",
			input: &TimeZone{
				Value: "--08:30",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[--08:30]"),
		},
		{
			name: "No ':': +0830",
			input: &TimeZone{
				Value: "+0830",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[+0830]"),
		},
		{
			name: "Invalid time format: +25:30",
			input: &TimeZone{
				Value: "+25:30",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[+25:30] (hr: 0~12)"),
		},
		{
			name: "Invalid time format: -00:60",
			input: &TimeZone{
				Value: "-00:60",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[-00:60] (min: 0~59)"),
		},
		{
			name: "Invalid DST format: -00:15-1",
			input: &TimeZone{
				Value: "-00:15-1",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[-00:15-1]"),
		},
		{
			name: "Invalid DST format: -00:15+3",
			input: &TimeZone{
				Value: "-00:15+3",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[-00:15+3]"),
		},
		{
			name: "Invalid DST format: -00:15+",
			input: &TimeZone{
				Value: "-00:15+",
			},
			expected: errors.Errorf("TimeZone MarshalBinary(): invalid format[-00:15+]"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.input.MarshalBinary()
			require.EqualError(t, err, tc.expected.Error())
		})
	}
}
