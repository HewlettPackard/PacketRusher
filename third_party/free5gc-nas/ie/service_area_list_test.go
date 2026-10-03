package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSvcAreaListUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *SvcAreaList
	}{
		{
			name:  "Positive Case 0 - all allowed",
			input: []byte{0x60, 0x13, 0x00, 0x13},
			expected: &SvcAreaList{
				AllAllowed: true,
			},
		},
		{
			name:  "Positive Case 1 - non-consec",
			input: []byte{0x80, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &SvcAreaList{
				DeniedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}},
				},
			},
		},
		{
			name:  "Positive Case 1.1 - non-consec",
			input: []byte{0x01, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56},
			expected: &SvcAreaList{
				AllowedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}, {
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "123456",
					}},
				},
			},
		},
		{
			name:  "Positive Case 2 - consec",
			input: []byte{0x20, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &SvcAreaList{
				AllowedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}},
				},
			},
		},
		{
			name:  "Positive Case 2.1 - consec, 2 elements",
			input: []byte{0x21, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &SvcAreaList{
				AllowedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}, {
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000002",
					}},
				},
			},
		},
		{
			name:  "Positive Case 3 - diff PlmnId, 1 element",
			input: []byte{0x40, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			expected: &SvcAreaList{
				AllowedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}},
				},
			},
		},
		{
			name:  "Positive Case 3.1 - diff PlmnId, 2 elements",
			input: []byte{0x41, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x00, 0x00, 0x02},
			expected: &SvcAreaList{
				AllowedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}, {
						PlmnId: PlmnId{MCC: "214", MNC: "653"},
						TAC:    "000002",
					}},
				},
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(SvcAreaList)
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

func TestSvcAreaListMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *SvcAreaList
		expected []byte
	}{
		// FIX-ME: how to handle AllAllowed?
		// {
		// 	name:     "Positive Case 0 - all allowed",
		// 	expected: []byte{0x60, 0x00, 0x00, 0x00},
		// 	input: &SvcAreaList{
		// 		AllAllowed: true,
		// 	},
		// },
		{
			name:     "Positive Case 1",
			expected: []byte{0x80, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01},
			input: &SvcAreaList{
				DeniedList: TrackingAreaIdList5GS{
					TAI: []TrackingAreaId5GS{{
						PlmnId: PlmnId{MCC: "310", MNC: "310"},
						TAC:    "000001",
					}},
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
