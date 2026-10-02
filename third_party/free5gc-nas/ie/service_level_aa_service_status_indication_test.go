package ie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSvcLvlAASvcStatusInd_UnmarshalBinary(t *testing.T) {
	tests := []struct {
		name   string
		b      []byte
		expUas bool
		expErr bool
	}{
		{
			name:   "Normal case 1",
			b:      []byte{0x01},
			expUas: true,
			expErr: false,
		},
		{
			name:   "Normal case 2",
			b:      []byte{0x31},
			expUas: true,
			expErr: false,
		},
		{
			name:   "Invalid input",
			b:      []byte{},
			expUas: false,
			expErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &SvcLvlAASvcStatusInd{}
			err := i.UnmarshalBinary(tt.b)
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expUas, i.uas)
			}
		})
	}
}

func TestSvcLvlAASvcStatusInd_MarshalBinary(t *testing.T) {
	tests := []struct {
		name   string
		input  *SvcLvlAASvcStatusInd
		expB   []byte
		expErr bool
	}{
		{
			name: "True case",
			input: &SvcLvlAASvcStatusInd{
				uas: true,
			},
			expB:   []byte{0x01},
			expErr: false,
		},
		{
			name: "False case",
			input: &SvcLvlAASvcStatusInd{
				uas: false,
			},
			expB:   []byte{0x00},
			expErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.input.MarshalBinary()
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expB, b)
			}
		})
	}
}
