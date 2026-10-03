package ie

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtendedDRXParams_UnmarshalBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   []byte
		want    ExtendedDRXParams
		wantErr bool
	}{
		{
			name:  "valid input with NR paging time window 1.28s and eDRX value 2.56s",
			input: []byte{0x00}, // PagingTimeWindow=0x00, eDRXValue=0x00
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue2_56s,
			},
			wantErr: false,
		},
		{
			name:  "valid input with NR paging time window 20.48s and eDRX value 10485.76s",
			input: []byte{0xf0}, // PagingTimeWindow=0x0f, eDRXValue=0x00
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow20_48s,
				EDRXValue:        NReDRXValue2_56s,
			},
			wantErr: false,
		},
		{
			name:  "valid input with eDRX value 5.12s",
			input: []byte{0x01}, // PagingTimeWindow=0x00, eDRXValue=0x01
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue5_12s,
			},
			wantErr: false,
		},
		{
			name:  "valid input with eDRX value 10.24s",
			input: []byte{0x02}, // PagingTimeWindow=0x00, eDRXValue=0x02
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue10_24s,
			},
			wantErr: false,
		},
		{
			name:  "valid input with eDRX value 10485.76s",
			input: []byte{0x0c}, // PagingTimeWindow=0x00, eDRXValue=0x0c
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue10485_76s,
			},
			wantErr: false,
		},
		{
			name:  "valid input with paging time window 8.96s and eDRX value 81.92s",
			input: []byte{0x65}, // PagingTimeWindow=0x06, eDRXValue=0x05
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow8_96s,
				EDRXValue:        NReDRXValue81_92s,
			},
			wantErr: false,
		},
		{
			name:  "invalid eDRX value (greater than max) should default to 2.56s",
			input: []byte{0x0d}, // PagingTimeWindow=0x00, eDRXValue=0x0d (invalid)
			want: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue2_56s, // Should default to this
			},
			wantErr: false,
		},
		{
			name:    "invalid buffer length (empty)",
			input:   []byte{},
			want:    ExtendedDRXParams{},
			wantErr: true,
		},
		{
			name:    "invalid buffer length (too long)",
			input:   []byte{0x00, 0x01},
			want:    ExtendedDRXParams{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ExtendedDRXParams
			err := got.UnmarshalBinary(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExtendedDRXParams_MarshalBinary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   ExtendedDRXParams
		want    []byte
		wantErr bool
	}{
		{
			name: "marshal NR paging time window 1.28s and eDRX value 2.56s",
			input: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue2_56s,
			},
			want:    []byte{0x00},
			wantErr: false,
		},
		{
			name: "marshal NR paging time window 20.48s and eDRX value 5.12s",
			input: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow20_48s,
				EDRXValue:        NReDRXValue5_12s,
			},
			want:    []byte{0xf1},
			wantErr: false,
		},
		{
			name: "marshal NR paging time window 10.24s and eDRX value 10.24s",
			input: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow10_24s,
				EDRXValue:        NReDRXValue10_24s,
			},
			want:    []byte{0x72},
			wantErr: false,
		},
		{
			name: "marshal NR paging time window 15.36s and eDRX value 10485.76s",
			input: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow15_36s,
				EDRXValue:        NReDRXValue10485_76s,
			},
			want:    []byte{0xbc},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalBinary()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExtendedDRXParams_MarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		orig ExtendedDRXParams
	}{
		{
			name: "round trip with NR paging time window 1.28s and eDRX value 2.56s",
			orig: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue2_56s,
			},
		},
		{
			name: "round trip with NR paging time window 20.48s and eDRX value 5.12s",
			orig: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow20_48s,
				EDRXValue:        NReDRXValue5_12s,
			},
		},
		{
			name: "round trip with NR paging time window 10.24s and eDRX value 10.24s",
			orig: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow10_24s,
				EDRXValue:        NReDRXValue10_24s,
			},
		},
		{
			name: "round trip with NR paging time window 15.36s and eDRX value 10485.76s",
			orig: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow15_36s,
				EDRXValue:        NReDRXValue10485_76s,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			marshaled, err := tt.orig.MarshalBinary()
			require.NoError(t, err)

			// Unmarshal
			var unmarshaled ExtendedDRXParams
			err = unmarshaled.UnmarshalBinary(marshaled)
			require.NoError(t, err)

			// Compare
			require.Equal(t, tt.orig, unmarshaled)
		})
	}
}

func TestTeDRXValueMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		eDRXValue NReDRXValue
		expected  NRTeDRXValue
	}{
		{
			name:      "NReDRXValue2_56s maps to NRTeDRXNotUsed",
			eDRXValue: NReDRXValue2_56s,
			expected:  NRTeDRXNotUsed,
		},
		{
			name:      "NReDRXValue5_12s maps to NRTeDRXNotUsed",
			eDRXValue: NReDRXValue5_12s,
			expected:  NRTeDRXNotUsed,
		},
		{
			name:      "NReDRXValue10_24s maps to NRTeDRXTwoPow0",
			eDRXValue: NReDRXValue10_24s,
			expected:  NRTeDRXTwoPow0,
		},
		{
			name:      "NReDRXValue20_48s maps to NRTeDRXTwoPow1",
			eDRXValue: NReDRXValue20_48s,
			expected:  NRTeDRXTwoPow1,
		},
		{
			name:      "NReDRXValue40_96s maps to NRTeDRXTwoPow2",
			eDRXValue: NReDRXValue40_96s,
			expected:  NRTeDRXTwoPow2,
		},
		{
			name:      "NReDRXValue81_92s maps to NRTeDRXTwoPow3",
			eDRXValue: NReDRXValue81_92s,
			expected:  NRTeDRXTwoPow3,
		},
		{
			name:      "NReDRXValue163_84s maps to NRTeDRXTwoPow4",
			eDRXValue: NReDRXValue163_84s,
			expected:  NRTeDRXTwoPow4,
		},
		{
			name:      "NReDRXValue327_68s maps to NRTeDRXTwoPow5",
			eDRXValue: NReDRXValue327_68s,
			expected:  NRTeDRXTwoPow5,
		},
		{
			name:      "NReDRXValue655_36s maps to NRTeDRXTwoPow6",
			eDRXValue: NReDRXValue655_36s,
			expected:  NRTeDRXTwoPow6,
		},
		{
			name:      "NReDRXValue1310_72s maps to NRTeDRXTwoPow7",
			eDRXValue: NReDRXValue1310_72s,
			expected:  NRTeDRXTwoPow7,
		},
		{
			name:      "NReDRXValue2621_44s maps to NRTeDRXTwoPow8",
			eDRXValue: NReDRXValue2621_44s,
			expected:  NRTeDRXTwoPow8,
		},
		{
			name:      "NReDRXValue5242_88s maps to NRTeDRXTwoPow9",
			eDRXValue: NReDRXValue5242_88s,
			expected:  NRTeDRXTwoPow9,
		},
		{
			name:      "NReDRXValue10485_76s maps to NRTeDRXTwoPow10",
			eDRXValue: NReDRXValue10485_76s,
			expected:  NRTeDRXTwoPow10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, exists := TeDRXValueMap[tt.eDRXValue]
			assert.True(t, exists, "eDRXValue %d should exist in TeDRXValueMap", tt.eDRXValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtendedDRXParams_Constants(t *testing.T) {
	// Test NR Paging Time Window constants
	require.Equal(t, NRPagingTimeWindow(0x00), NRPagingTimeWindow1_28s)
	require.Equal(t, NRPagingTimeWindow(0x01), NRPagingTimeWindow2_56s)
	require.Equal(t, NRPagingTimeWindow(0x02), NRPagingTimeWindow3_84s)
	require.Equal(t, NRPagingTimeWindow(0x03), NRPagingTimeWindow5_12s)
	require.Equal(t, NRPagingTimeWindow(0x04), NRPagingTimeWindow6_4s)
	require.Equal(t, NRPagingTimeWindow(0x05), NRPagingTimeWindow7_68s)
	require.Equal(t, NRPagingTimeWindow(0x06), NRPagingTimeWindow8_96s)
	require.Equal(t, NRPagingTimeWindow(0x07), NRPagingTimeWindow10_24s)
	require.Equal(t, NRPagingTimeWindow(0x08), NRPagingTimeWindow11_52s)
	require.Equal(t, NRPagingTimeWindow(0x09), NRPagingTimeWindow12_8s)
	require.Equal(t, NRPagingTimeWindow(0x0a), NRPagingTimeWindow14_08s)
	require.Equal(t, NRPagingTimeWindow(0x0b), NRPagingTimeWindow15_36s)
	require.Equal(t, NRPagingTimeWindow(0x0c), NRPagingTimeWindow16_64s)
	require.Equal(t, NRPagingTimeWindow(0x0d), NRPagingTimeWindow17_92s)
	require.Equal(t, NRPagingTimeWindow(0x0e), NRPagingTimeWindow19_2s)
	require.Equal(t, NRPagingTimeWindow(0x0f), NRPagingTimeWindow20_48s)

	// Test NR eDRX Value constants
	require.Equal(t, NReDRXValue(0x00), NReDRXValue2_56s)
	require.Equal(t, NReDRXValue(0x01), NReDRXValue5_12s)
	require.Equal(t, NReDRXValue(0x02), NReDRXValue10_24s)
	require.Equal(t, NReDRXValue(0x03), NReDRXValue20_48s)
	require.Equal(t, NReDRXValue(0x04), NReDRXValue40_96s)
	require.Equal(t, NReDRXValue(0x05), NReDRXValue81_92s)
	require.Equal(t, NReDRXValue(0x06), NReDRXValue163_84s)
	require.Equal(t, NReDRXValue(0x07), NReDRXValue327_68s)
	require.Equal(t, NReDRXValue(0x08), NReDRXValue655_36s)
	require.Equal(t, NReDRXValue(0x09), NReDRXValue1310_72s)
	require.Equal(t, NReDRXValue(0x0a), NReDRXValue2621_44s)
	require.Equal(t, NReDRXValue(0x0b), NReDRXValue5242_88s)
	require.Equal(t, NReDRXValue(0x0c), NReDRXValue10485_76s)

	// Test NR TeDRX constants
	require.Equal(t, NRTeDRXValue(0x00), NRTeDRXNotUsed)
	require.Equal(t, NRTeDRXValue(1), NRTeDRXTwoPow0)
	require.Equal(t, NRTeDRXValue(2), NRTeDRXTwoPow1)
	require.Equal(t, NRTeDRXValue(4), NRTeDRXTwoPow2)
	require.Equal(t, NRTeDRXValue(8), NRTeDRXTwoPow3)
	require.Equal(t, NRTeDRXValue(16), NRTeDRXTwoPow4)
	require.Equal(t, NRTeDRXValue(32), NRTeDRXTwoPow5)
	require.Equal(t, NRTeDRXValue(64), NRTeDRXTwoPow6)
	require.Equal(t, NRTeDRXValue(128), NRTeDRXTwoPow7)
	require.Equal(t, NRTeDRXValue(256), NRTeDRXTwoPow8)
	require.Equal(t, NRTeDRXValue(512), NRTeDRXTwoPow9)
	require.Equal(t, NRTeDRXValue(1024), NRTeDRXTwoPow10)
}

func TestNRPagingTimeWindow_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    NRPagingTimeWindow
		expected string
	}{
		{
			name:     "NRPagingTimeWindow1_28s",
			value:    NRPagingTimeWindow1_28s,
			expected: "1.28s",
		},
		{
			name:     "NRPagingTimeWindow2_56s",
			value:    NRPagingTimeWindow2_56s,
			expected: "2.56s",
		},
		{
			name:     "NRPagingTimeWindow3_84s",
			value:    NRPagingTimeWindow3_84s,
			expected: "3.84s",
		},
		{
			name:     "NRPagingTimeWindow5_12s",
			value:    NRPagingTimeWindow5_12s,
			expected: "5.12s",
		},
		{
			name:     "NRPagingTimeWindow6_4s",
			value:    NRPagingTimeWindow6_4s,
			expected: "6.4s",
		},
		{
			name:     "NRPagingTimeWindow7_68s",
			value:    NRPagingTimeWindow7_68s,
			expected: "7.68s",
		},
		{
			name:     "NRPagingTimeWindow8_96s",
			value:    NRPagingTimeWindow8_96s,
			expected: "8.96s",
		},
		{
			name:     "NRPagingTimeWindow10_24s",
			value:    NRPagingTimeWindow10_24s,
			expected: "10.24s",
		},
		{
			name:     "NRPagingTimeWindow11_52s",
			value:    NRPagingTimeWindow11_52s,
			expected: "11.52s",
		},
		{
			name:     "NRPagingTimeWindow12_8s",
			value:    NRPagingTimeWindow12_8s,
			expected: "12.8s",
		},
		{
			name:     "NRPagingTimeWindow14_08s",
			value:    NRPagingTimeWindow14_08s,
			expected: "14.08s",
		},
		{
			name:     "NRPagingTimeWindow15_36s",
			value:    NRPagingTimeWindow15_36s,
			expected: "15.36s",
		},
		{
			name:     "NRPagingTimeWindow16_64s",
			value:    NRPagingTimeWindow16_64s,
			expected: "16.64s",
		},
		{
			name:     "NRPagingTimeWindow17_92s",
			value:    NRPagingTimeWindow17_92s,
			expected: "17.92s",
		},
		{
			name:     "NRPagingTimeWindow19_2s",
			value:    NRPagingTimeWindow19_2s,
			expected: "19.2s",
		},
		{
			name:     "NRPagingTimeWindow20_48s",
			value:    NRPagingTimeWindow20_48s,
			expected: "20.48s",
		},
		{
			name:     "Unknown value",
			value:    NRPagingTimeWindow(0xff),
			expected: "Unknown(255)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNReDRXValue_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    NReDRXValue
		expected string
	}{
		{
			name:     "NReDRXValue2_56s",
			value:    NReDRXValue2_56s,
			expected: "Len:2.56s, TeDRX:NotUsed",
		},
		{
			name:     "NReDRXValue5_12s",
			value:    NReDRXValue5_12s,
			expected: "Len:5.12s, TeDRX:NotUsed",
		},
		{
			name:     "NReDRXValue10_24s",
			value:    NReDRXValue10_24s,
			expected: "Len:10.24s, TeDRX:2^0=1",
		},
		{
			name:     "NReDRXValue20_48s",
			value:    NReDRXValue20_48s,
			expected: "Len:20.48s, TeDRX:2^1=2",
		},
		{
			name:     "NReDRXValue40_96s",
			value:    NReDRXValue40_96s,
			expected: "Len:40.96s, TeDRX:2^2=4",
		},
		{
			name:     "NReDRXValue81_92s",
			value:    NReDRXValue81_92s,
			expected: "Len:81.92s, TeDRX:2^3=8",
		},
		{
			name:     "NReDRXValue163_84s",
			value:    NReDRXValue163_84s,
			expected: "Len:163.84s, TeDRX:2^4=16",
		},
		{
			name:     "NReDRXValue327_68s",
			value:    NReDRXValue327_68s,
			expected: "Len:327.68s, TeDRX:2^5=32",
		},
		{
			name:     "NReDRXValue655_36s",
			value:    NReDRXValue655_36s,
			expected: "Len:655.36s, TeDRX:2^6=64",
		},
		{
			name:     "NReDRXValue1310_72s",
			value:    NReDRXValue1310_72s,
			expected: "Len:1310.72s, TeDRX:2^7=128",
		},
		{
			name:     "NReDRXValue2621_44s",
			value:    NReDRXValue2621_44s,
			expected: "Len:2621.44s, TeDRX:2^8=256",
		},
		{
			name:     "NReDRXValue5242_88s",
			value:    NReDRXValue5242_88s,
			expected: "Len:5242.88s, TeDRX:2^9=512",
		},
		{
			name:     "NReDRXValue10485_76s",
			value:    NReDRXValue10485_76s,
			expected: "Len:10485.76s, TeDRX:2^10=1024",
		},
		{
			name:     "Unknown value",
			value:    NReDRXValue(0xff),
			expected: "Unknown(255)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNRTeDRXValue_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    NRTeDRXValue
		expected string
	}{
		{
			name:     "NRTeDRXNotUsed",
			value:    NRTeDRXNotUsed,
			expected: "NotUsed",
		},
		{
			name:     "NRTeDRXTwoPow0",
			value:    NRTeDRXTwoPow0,
			expected: "2^0=1",
		},
		{
			name:     "NRTeDRXTwoPow1",
			value:    NRTeDRXTwoPow1,
			expected: "2^1=2",
		},
		{
			name:     "NRTeDRXTwoPow2",
			value:    NRTeDRXTwoPow2,
			expected: "2^2=4",
		},
		{
			name:     "NRTeDRXTwoPow3",
			value:    NRTeDRXTwoPow3,
			expected: "2^3=8",
		},
		{
			name:     "NRTeDRXTwoPow4",
			value:    NRTeDRXTwoPow4,
			expected: "2^4=16",
		},
		{
			name:     "NRTeDRXTwoPow5",
			value:    NRTeDRXTwoPow5,
			expected: "2^5=32",
		},
		{
			name:     "NRTeDRXTwoPow6",
			value:    NRTeDRXTwoPow6,
			expected: "2^6=64",
		},
		{
			name:     "NRTeDRXTwoPow7",
			value:    NRTeDRXTwoPow7,
			expected: "2^7=128",
		},
		{
			name:     "NRTeDRXTwoPow8",
			value:    NRTeDRXTwoPow8,
			expected: "2^8=256",
		},
		{
			name:     "NRTeDRXTwoPow9",
			value:    NRTeDRXTwoPow9,
			expected: "2^9=512",
		},
		{
			name:     "NRTeDRXTwoPow10",
			value:    NRTeDRXTwoPow10,
			expected: "2^10=1024",
		},
		{
			name:     "Unknown value",
			value:    NRTeDRXValue(0xffff),
			expected: "Unknown(65535)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtendedDRXParams_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		params   ExtendedDRXParams
		expected string
	}{
		{
			name: "basic params",
			params: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow1_28s,
				EDRXValue:        NReDRXValue2_56s,
			},
			expected: "PagingTimeWindow: 1.28s; eDRXValue: Len:2.56s, TeDRX:NotUsed",
		},
		{
			name: "maximum values",
			params: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow20_48s,
				EDRXValue:        NReDRXValue10485_76s,
			},
			expected: "PagingTimeWindow: 20.48s; eDRXValue: Len:10485.76s, TeDRX:2^10=1024",
		},
		{
			name: "mixed values",
			params: ExtendedDRXParams{
				PagingTimeWindow: NRPagingTimeWindow10_24s,
				EDRXValue:        NReDRXValue655_36s,
			},
			expected: "PagingTimeWindow: 10.24s; eDRXValue: Len:655.36s, TeDRX:2^6=64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
