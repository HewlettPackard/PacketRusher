package ie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessAMBRUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *SessAMBR
	}{
		{
			name:  "Positive Case 1",
			input: []byte{0x16, 0x12, 0x34, 0x0f, 0x56, 0x78},
			expected: &SessAMBR{
				UnitDownlink:  Rate_4Pbps,
				ValueDownlink: 0x1234,
				UnitUplink:    Rate_256Gbps,
				ValueUplink:   0x5678,
			},
		},
		{
			name:  "Negative Case 1",
			input: []byte{0x76, 0x54},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(SessAMBR)
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

func TestSessAMBRMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *SessAMBR
		expected []byte
	}{
		{
			name: "Positive Case 1",
			input: &SessAMBR{
				UnitDownlink:  Rate_4Pbps,
				ValueDownlink: 0x1234,
				UnitUplink:    Rate_256Gbps,
				ValueUplink:   0x5678,
			},
			expected: []byte{0x16, 0x12, 0x34, 0x0f, 0x56, 0x78},
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

func TestSessAMBR_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ul       string
		dl       string
		expected SessAMBR
	}{
		{
			name: "normal Pbps / lower case Kbps",
			ul:   "999 Pbps",
			dl:   "3 Kbps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Pbps,
				ValueUplink:   999,
				UnitDownlink:  Rate_1Kbps,
				ValueDownlink: 0x3,
			},
		},
		{
			name: "Mbps, Tbps",
			ul:   "1024 Mbps",
			dl:   "333 Tbps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Mbps,
				ValueUplink:   1024,
				UnitDownlink:  Rate_1Tbps,
				ValueDownlink: 333,
			},
		},
		{
			name: "Default Unit, Gbps, MAX of Uint16",
			ul:   "10 sdfMsdfbps",
			dl:   "65535 Gbps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Kbps,
				ValueUplink:   10,
				UnitDownlink:  Rate_1Gbps,
				ValueDownlink: 0xffff,
			},
		},
		{
			name: ">= 65535",
			ul:   "65535 Kbps",
			dl:   "65536 Kbps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Kbps,
				ValueUplink:   65535,
				UnitDownlink:  Rate_4Kbps,
				ValueDownlink: 0x4000,
			},
		},
		{
			name: "much >= 65535",
			ul:   "10000000 bps",
			dl:   "10000000000 bps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Kbps,
				ValueUplink:   0x2710,
				UnitDownlink:  Rate_256Kbps,
				ValueDownlink: 0x9896,
			},
		},
		{
			name: "smaller bps",
			ul:   "1 bps",
			dl:   "999 bps",
			expected: SessAMBR{
				UnitUplink:    Rate_1Kbps,
				ValueUplink:   0x0,
				UnitDownlink:  Rate_1Kbps,
				ValueDownlink: 0x0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := SessAMBR{}
			err := s.Set(tc.ul, tc.dl)
			require.NoError(t, err)
			require.Equal(t, tc.expected, s)
		})
	}
}

func TestBriefSessAMBR(t *testing.T) {
	testCases := []struct {
		name string
		ambr *SessAMBR
		want string
	}{
		{
			name: "4K",
			ambr: &SessAMBR{
				ValueUplink:   10,
				UnitUplink:    Rate_4Kbps,
				ValueDownlink: 20,
				UnitDownlink:  Rate_4Kbps,
			},
			want: "SessAMBR[UL:10*4K,DL:20*4K]",
		},
		{
			name: "1M",
			ambr: &SessAMBR{
				ValueUplink:   100,
				UnitUplink:    Rate_1Mbps,
				ValueDownlink: 200,
				UnitDownlink:  Rate_1Mbps,
			},
			want: "SessAMBR[UL:100*1M,DL:200*1M]",
		},
		{
			name: "16Gbps",
			ambr: &SessAMBR{
				ValueUplink:   30,
				UnitUplink:    Rate_16Gbps,
				ValueDownlink: 40,
				UnitDownlink:  Rate_16Gbps,
			},
			want: "SessAMBR[UL:30*16G,DL:40*16G]",
		},
		{
			name: "256Tbps",
			ambr: &SessAMBR{
				ValueUplink:   50,
				UnitUplink:    Rate_256Tbps,
				ValueDownlink: 60,
				UnitDownlink:  Rate_256Tbps,
			},
			want: "SessAMBR[UL:50*256T,DL:60*256T]",
		},
		{
			name: "64Pbps",
			ambr: &SessAMBR{
				ValueUplink:   70,
				UnitUplink:    Rate_64Pbps,
				ValueDownlink: 80,
				UnitDownlink:  Rate_64Pbps,
			},
			want: "SessAMBR[UL:70*64P,DL:80*64P]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.ambr.String())
		})
	}
}
