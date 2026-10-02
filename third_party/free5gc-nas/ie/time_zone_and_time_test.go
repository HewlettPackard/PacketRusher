package ie

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeZoneAndTimeUnmarshalBinary(t *testing.T) {
	CST, err := time.LoadLocation("Asia/Taipei")
	require.NoError(t, err)
	IST, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)
	EST, err := time.LoadLocation("America/New_York") // If using DST, EST would be EDT.
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    []byte
		expected *TimeZoneAndTime
	}{
		{
			// CST = UTC+8
			name: "2023-7-13 12:27:39 CST",
			input: []byte{
				0x32, 0x70, 0x31, 0x21, 0x72, 0x93, 0x23,
			},
			expected: &TimeZoneAndTime{
				Time: time.Date(2023, time.July, 13, 12, 27, 39, 0, CST),
			},
		},
		{
			// IST = UTC-5:30
			name: "2019-12-15 16:55:46 IST",
			input: []byte{
				0x91, 0x21, 0x51, 0x61, 0x55, 0x64, 0x22,
			},
			expected: &TimeZoneAndTime{
				Time: time.Date(2019, time.December, 15, 16, 55, 46, 0, IST),
			},
		},
		{
			// EST = UTC-5 (not DST)
			name: "2001-2-2 09:03:06 EST",
			input: []byte{
				0x10, 0x20, 0x20, 0x90, 0x30, 0x60, 0x0A,
			},
			expected: &TimeZoneAndTime{
				Time: time.Date(2001, time.February, 2, 9, 3, 6, 0, EST),
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tzAndTime := new(TimeZoneAndTime)
			err := tzAndTime.UnmarshalBinary(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected.Time.Format(time.RFC822Z), tzAndTime.Time.Format(time.RFC822Z))
		})
	}
}

func TestTimeZoneAndTimeMarshalBinary(t *testing.T) {
	CST, err := time.LoadLocation("Asia/Taipei")
	require.NoError(t, err)
	IST, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)
	EST, err := time.LoadLocation("America/New_York") // If using DST, EST would be EDT.
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    *TimeZoneAndTime
		expected []byte
	}{
		{
			// CST = UTC+8
			name: "2023-7-13 12:27:39 CST",
			input: &TimeZoneAndTime{
				Time: time.Date(2023, time.July, 13, 12, 27, 39, 0, CST),
			},
			expected: []byte{
				0x32, 0x70, 0x31, 0x21, 0x72, 0x93, 0x23,
			},
		},
		{
			// IST = UTC-5:30
			name: "2019-12-15 16:55:46 IST",
			input: &TimeZoneAndTime{
				Time: time.Date(2019, time.December, 15, 16, 55, 46, 0, IST),
			},
			expected: []byte{
				0x91, 0x21, 0x51, 0x61, 0x55, 0x64, 0x22,
			},
		},
		{
			// EST = UTC-5 (not DST)
			name: "2001-2-2 09:03:06 EST",
			input: &TimeZoneAndTime{
				Time: time.Date(2001, time.February, 2, 9, 3, 6, 0, EST),
			},
			expected: []byte{
				0x10, 0x20, 0x20, 0x90, 0x30, 0x60, 0x0A,
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bytes, err := tc.input.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.expected, bytes)
		})
	}
}
