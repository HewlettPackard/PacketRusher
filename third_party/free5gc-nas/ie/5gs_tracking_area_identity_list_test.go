package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrackingAreaIdList5GSUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *TrackingAreaIdList5GS
	}{
		{
			name:  "Positive Case 1 - non-consec",
			input: []byte{0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}},
			},
		},
		{
			name:  "Positive Case 1.1 - non-consec",
			input: []byte{0x01, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "123456",
				}},
			},
		},
		{
			name:  "Positive Case 2 - consec",
			input: []byte{0x20, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}},
			},
		},
		{
			name:  "Positive Case 2.1 - consec, 2 elements",
			input: []byte{0x21, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000002",
				}},
			},
		},
		{
			name:  "Positive Case 3 - diff PlmnId, 1 element",
			input: []byte{0x40, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}},
			},
		},
		{
			name:  "Positive Case 3.1 - diff PlmnId, 2 elements",
			input: []byte{0x41, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x00, 0x00, 0x02},
			expected: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "214", MNC: "653"},
					TAC:    "000002",
				}},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(TrackingAreaIdList5GS)
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

func TestTrackingAreaIdList5GSMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *TrackingAreaIdList5GS
		expected []byte
	}{
		{
			name:     "1 item in the list",
			expected: []byte{0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			input: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}},
			},
		},
		{
			name: "3 Consecutive",
			expected: []byte{
				0x22,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x02,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x03,
			},
			input: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000002",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000003",
				}},
			},
		},
		{
			name: "2 Consecutive + 1 non-consecutive",
			expected: []byte{
				0x02,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x02,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x04,
			},
			input: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000002",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000004",
				}},
			},
		},
		{
			name: "Different PLMN, Consecutive TAC",
			expected: []byte{
				0x42,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
				0x13, 0x00, 0x13, 0x00, 0x00, 0x02,
				0x14, 0x00, 0x14, 0x00, 0x00, 0x03,
			},
			input: &TrackingAreaIdList5GS{
				TAI: []TrackingAreaId5GS{{
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000001",
				}, {
					PlmnId: PlmnId{MCC: "310", MNC: "310"},
					TAC:    "000002",
				}, {
					PlmnId: PlmnId{MCC: "410", MNC: "410"},
					TAC:    "000003",
				}},
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
