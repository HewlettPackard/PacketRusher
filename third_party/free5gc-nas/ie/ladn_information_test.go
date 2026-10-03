package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLADNInfoUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *LADNInfo
	}{
		{
			name: "Positive Case 1 - non-consec",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07,                                     // TAI Length
				0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, // TAI
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 1.1 - non-consec",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x0A, // TAI Length
				0x01, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56,
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "123456",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 2 - consec",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07, // TAI Length
				0x20, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 2.1 - consec, 2 elements",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07, // TAI Length
				0x21, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000002",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 3 - diff PlmnId, 1 element",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07, // TAI Length
				0x40, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 3.1 - diff PlmnId, 2 elements",
			input: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x0D, // TAI Length
				0x41, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x00, 0x00, 0x02,
			},
			expected: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "214", MNC: "653"},
							TAC:    "000002",
						}},
					},
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
			ie := new(LADNInfo)
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

func TestLADNInfoMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *LADNInfo
		expected []byte
	}{
		{
			name: "Positive Case 1 - non-consec",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07,                                     // TAI Length
				0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, // TAI
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 1.1 - non-consec",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x0d, // TAI Length
				0x01, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x13, 0x00, 0x13, 0x12, 0x34, 0x56,
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "123456",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 2 - consec",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07, // TAI Length
				0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 2.1 - consec, 2 elements",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x0d, // TAI Length
				0x21, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x13, 0x00, 0x13, 0x00, 0x00, 0x02,
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000002",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 3 - diff PlmnId, 1 element",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x07, // TAI Length
				0x00, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01,
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}},
					},
				}},
			},
		},
		{
			name: "Positive Case 3.1 - diff PlmnId, 2 elements",
			expected: []byte{
				0x04,                   // DNN Length
				0x03, 0x61, 0x62, 0x63, // DNN
				0x0D, // TAI Length
				0x41, 0x13, 0x00, 0x13, 0x00, 0x00, 0x01, 0x12, 0x34, 0x56, 0x00, 0x00, 0x02,
			},
			input: &LADNInfo{
				DnnTai: []DNN_TAI{{
					DNN: DNN{
						Value: "abc",
					},
					TrackingAreaIdList5GS: TrackingAreaIdList5GS{
						TAI: []TrackingAreaId5GS{{
							PlmnId: PlmnId{MCC: "310", MNC: "310"},
							TAC:    "000001",
						}, {
							PlmnId: PlmnId{MCC: "214", MNC: "653"},
							TAC:    "000002",
						}},
					},
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
