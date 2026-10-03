package ie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQosFlowDescsUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *QosFlowDescs
	}{
		{
			name:  "Positive Case - Create 5QI",
			input: []byte{0x01, 0x20, 0x41, 0x01, 0x01, 0x08},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    1,
						OpCode: QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(Qfd_5QI8),
					},
				},
			},
		},
		{
			name:  "Positive Case - Create GFBR uplink 100Mbps",
			input: []byte{0x01, 0x20, 0x41, 0x02, 0x03, 0x06, 0x00, 0x64},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:        1,
						OpCode:     QFD_Create,
						EBit:       1, // parameters list is included
						GFBRUplink: "100 Mbps",
					},
				},
			},
		},
		{
			name: "Positive Case - Create GFBR/MFBR UL/DL",
			input: []byte{
				0x01, 0x20, 0x44,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
			},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Create,
						EBit:         1, // parameters list is included
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "100 Mbps",
						MFBRUplink:   "100 Gbps",
						MFBRDownlink: "100 Tbps",
					},
				},
			},
		},
		{
			name: "Positive Case - Create GFBR/MFBR UL/DL, AvgWin, 3 * EPSBearerId",
			input: []byte{
				0x01, 0x20, 0x48,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
				0x06, 0x02, 0x00, 0x64,
				0x07, 0x01, 0x70,
				0x07, 0x01, 0x20,
				0x07, 0x01, 0x50,
			},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Create,
						EBit:         1, // parameters list is included
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "100 Mbps",
						MFBRUplink:   "100 Gbps",
						MFBRDownlink: "100 Tbps",
						AvgWin:       100,
						EPSBearerIds: QFP_EPSBearerId{
							NumOfBearerId: 3,
							Id: [15]byte{
								7, 2, 5, 0, 0,
								0, 0, 0, 0, 0,
								0, 0, 0, 0, 0,
							},
						},
					},
				},
			},
		},
		{
			name:  "Positive Case - Del existing QFD",
			input: []byte{0x01, 0x40, 0x00},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    1,
						OpCode: QFD_Del,
						EBit:   0, // parameters list is not included
					},
				},
			},
		},
		{
			name:  "Negative Case - Del existing QFD, # of params > 0",
			input: []byte{0x01, 0x40, 0x05},
		},
		{
			name: "Negative Case - Del existing QFD but Ebit is 1",
			input: []byte{
				0x01, 0x40, 0x40, // 0x01, 0x40, 0x00,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
		{
			name: "Negative Case - Create 5QI, bad # of params",
			input: []byte{
				0x01, 0x20, 0x42, // 01 20 41
				0x01, 0x01, 0x08,
			},
		},
		{
			name: "Negative Case - Create 5QI, bad len of param content",
			input: []byte{
				0x01, 0x20, 0x41,
				0x01, 0x02, 0x08, // 01 01 08
			},
		},
		{
			name: "Negative Case - Create GFBR uplink 100Mbps, bad # of params",
			input: []byte{
				0x01, 0x20, 0x40, // 0x01, 0x20, 0x41,
				0x02, 0x03, 0x06, 0x00, 0x64,
			},
		},
		{
			name: "Negative Case - Create GFBR uplink 100Mbps, bad len of prarm content",
			input: []byte{
				0x01, 0x20, 0x41,
				0x02, 0x04, 0x06, 0x00, 0x64, // 0x02, 0x03, 0x06, 0x00, 0x64,
			},
		},
		{
			name: "Negative Case - Create GFBR/MFBR UL/DL, AvgWin, 3 * EPSBearerId, bad # of param content",
			input: []byte{
				0x01, 0x20, 0x4A, // 0x01, 0x20, 0x48,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
				0x06, 0x02, 0x00, 0x64,
				0x07, 0x01, 0x70,
				0x07, 0x01, 0x20,
				0x07, 0x01, 0x50,
			},
		},
		{
			name: "QoS flow description * 2",
			input: []byte{
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
				0x09, 0x20, 0x41, 0x01, 0x01, 0x09,
			},
			expected: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    9,
						OpCode: QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(Qfd_5QI9),
					},
					{
						QFI:    9,
						OpCode: QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(Qfd_5QI9),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(QosFlowDescs)
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

func TestQosFlowDescsMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *QosFlowDescs
		expected []byte
	}{
		{
			name:     "Positive Case 1 - Create 5QI",
			expected: []byte{0x01, 0x20, 0x41, 0x01, 0x01, 0x09},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    1,
						OpCode: QFD_Create,
						EBit:   1, // parameters list is included
						FiveQI: uint8(Qfd_5QI9),
					},
				},
			},
		},
		{
			name:     "Positive Case 2 - Del existing QFD",
			expected: []byte{0x01, 0x40, 0x00},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    1,
						OpCode: QFD_Del,
						EBit:   0, // parameters list is not included
					},
				},
			},
		},
		{
			name:     "Positive Case - Create GFBR uplink 100Mbps",
			expected: []byte{0x01, 0x20, 0x41, 0x02, 0x03, 0x06, 0x00, 0x64},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:        1,
						OpCode:     QFD_Create,
						EBit:       1, // parameters list is included
						GFBRUplink: "100 Mbps",
					},
				},
			},
		},
		{
			name: "Positive Case - Create GFBR/MFBR UL/DL, AvgWin, 3 * EPSBearerId",
			expected: []byte{
				0x01, 0x20, 0x48,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
				0x06, 0x02, 0x00, 0x64,
				0x07, 0x01, 0x70,
				0x07, 0x01, 0x20,
				0x07, 0x01, 0x50,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Create,
						EBit:         1, // parameters list is included
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "100 Mbps",
						MFBRUplink:   "100 Gbps",
						MFBRDownlink: "100 Tbps",
						AvgWin:       100,
						EPSBearerIds: QFP_EPSBearerId{
							NumOfBearerId: 3,
							Id: [15]byte{
								7, 2, 5, 0, 0,
								0, 0, 0, 0, 0,
								0, 0, 0, 0, 0,
							},
						},
					},
				},
			},
		},
		{
			name: "Positive Case - Create GFBR/MFBR UL/DL",
			expected: []byte{
				0x01, 0x20, 0x44,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:    1,
						OpCode: QFD_Create,
						//		EBit:         1, // parameters list is included
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "100 Mbps",
						MFBRUplink:   "100 Gbps",
						MFBRDownlink: "100 Tbps",
					},
				},
			},
		},
		{
			name: "Positive Case - Modify GFBR/MFBR UL/DL",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x01, 0x00, 0x64,
				0x03, 0x03, 0x06, 0x00, 0x64,
				0x04, 0x03, 0x0b, 0x00, 0x64,
				0x05, 0x03, 0x10, 0x00, 0x64,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "100 Mbps",
						MFBRUplink:   "100 Gbps",
						MFBRDownlink: "100 Tbps",
					},
				},
			},
		},
		{
			name: "Positive Case - >= 0xFFFF units",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x01, 0xff, 0xff,
				0x03, 0x03, 0x07, 0x40, 0x00,
				0x04, 0x03, 0x0b, 0xFF, 0xFF,
				0x05, 0x03, 0x10, 0xFF, 0xFF,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "65535 Kbps",
						GFBRDownlink: "65536 Mbps",
						MFBRUplink:   "65535 Gbps",
						MFBRDownlink: "65535 Tbps",
					},
				},
			},
		},
		{
			name: "Positive Case - way more than 0xFFFF units",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x05, 0x98, 0x96,
				0x03, 0x03, 0x08, 0xF4, 0x24,
				0x04, 0x03, 0x0c, 0x61, 0xA8,
				0x05, 0x03, 0x11, 0xFF, 0xFF,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "10000000 Kbps",
						GFBRDownlink: "1000000 Mbps",
						MFBRUplink:   "100000 Gbps",
						MFBRDownlink: "262140 Tbps",
					},
				},
			},
		},
		{
			name: "Positive Case - bps",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x01, 0x00, 0x01,
				0x03, 0x03, 0x01, 0x00, 0x0A,
				0x04, 0x03, 0x01, 0x00, 0x64,
				0x05, 0x03, 0x01, 0x03, 0xE8,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "1000 bps",    //    1Kbps
						GFBRDownlink: "10000 bps",   //   10Kbps
						MFBRUplink:   "100000 bps",  //  100Kbps
						MFBRDownlink: "1000000 bps", // 1000Kbps
					},
				},
			},
		},
		{
			name: "Positive Case - larger bps",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x01, 0x27, 0x10,
				0x03, 0x03, 0x02, 0x61, 0xA8,
				0x04, 0x03, 0x03, 0xF4, 0x24,
				0x05, 0x03, 0x05, 0x98, 0x96,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "10000000 bps",    //    10000Kbps
						GFBRDownlink: "100000000 bps",   //   100000Kbps
						MFBRUplink:   "1000000000 bps",  //  1000000Kbps
						MFBRDownlink: "10000000000 bps", // 10000000Kbps
					},
				},
			},
		},
		{
			name: "Positive Case - smaller bps",
			expected: []byte{
				0x01, 0x60, 0x44,
				0x02, 0x03, 0x01, 0x00, 0x00,
				0x03, 0x03, 0x01, 0x00, 0x00,
				0x04, 0x03, 0x01, 0x00, 0x00,
				0x05, 0x03, 0x01, 0x00, 0x00,
			},
			input: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						QFI:          1,
						OpCode:       QFD_Mod,
						GFBRUplink:   "1 bps",
						GFBRDownlink: "10 bps",
						MFBRUplink:   "100 bps",
						MFBRDownlink: "999 bps",
					},
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

func TestBriefNasQfd(t *testing.T) {
	testCases := []struct {
		name   string
		nasQfd *QosFlowDescs
		want   string
	}{
		{
			name:   "case 1",
			nasQfd: &QosFlowDescs{},
			want:   "",
		},
		{
			name: "case 2",
			nasQfd: &QosFlowDescs{
				Descs: []QosFlowDesc{},
			},
			want: "",
		},
		{
			name: "case 3",
			nasQfd: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{},
				},
			},
			want: "",
		},
		{
			name: "case 4",
			nasQfd: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{},
					{
						// Del QFD should only have QFI, but printing all of
						// them helps detect if bad values have sent out.
						OpCode:       QFD_Del,
						QFI:          1,
						FiveQI:       2,
						GFBRUplink:   "10 Mbps",
						GFBRDownlink: "20 Mbps",
						MFBRUplink:   "30 Mbps",
						MFBRDownlink: "40 Mbps",
					},
					{},
				},
			},
			want: "del QFD[QFI:1,5QI:2,GBR(UL:10M,DL:20M),MBR(UL:30M,DL:40M)]",
		},
		{
			name: "case 5",
			nasQfd: &QosFlowDescs{
				Descs: []QosFlowDesc{
					{
						OpCode:       QFD_Create,
						QFI:          1,
						FiveQI:       2,
						GFBRUplink:   "10 Mbps",
						GFBRDownlink: "20 Mbps",
						MFBRUplink:   "30 Mbps",
						MFBRDownlink: "40 Mbps",
					},
					{
						OpCode:       QFD_Del,
						QFI:          2,
						FiveQI:       0,
						GFBRUplink:   "",
						GFBRDownlink: "",
						MFBRUplink:   "",
						MFBRDownlink: "",
					},
					{
						OpCode:       QFD_Mod,
						QFI:          3,
						FiveQI:       4,
						GFBRUplink:   "100 Kbps",
						GFBRDownlink: "200 Kbps",
						MFBRUplink:   "300 Kbps",
						MFBRDownlink: "400 Kbps",
					},
					{
						OpCode:       QFD_Create,
						QFI:          4,
						FiveQI:       9,
						GFBRUplink:   "",
						GFBRDownlink: "",
						MFBRUplink:   "",
						MFBRDownlink: "",
					},
				},
			},

			want: "add QFD[QFI:1,5QI:2,GBR(UL:10M,DL:20M),MBR(UL:30M,DL:40M)]," +
				"del QFD[QFI:2]," +
				"mod QFD[QFI:3,5QI:4,GBR(UL:100K,DL:200K),MBR(UL:300K,DL:400K)]," +
				"add QFD[QFI:4,5QI:9]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.nasQfd.String())
		})
	}
}
