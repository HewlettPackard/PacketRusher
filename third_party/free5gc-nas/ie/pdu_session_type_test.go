package ie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPduSessionTypeUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		ie    PDUSessType
		input []byte
		ieOk  PDUSessType
		out   []byte
	}{
		{
			name:  "",
			ie:    PDUSessType{},
			input: []byte{0x2E},
			ieOk: PDUSessType{
				Value: 6,
			},
			out: []byte{0x06},
		},
		{
			name:  "",
			ie:    PDUSessType{},
			input: []byte{0xFE},
			ieOk: PDUSessType{
				Value: 6,
			},
			out: []byte{0x06},
		},
		{
			name:  "",
			ie:    PDUSessType{},
			input: []byte{0xA2},
			ieOk: PDUSessType{
				Value: 2,
			},
			out: []byte{0x02},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out []byte
			var err error

			if err = tc.ie.UnmarshalBinary(tc.input); err != nil {
				t.Error(err)
			}
			assert.Equal(t, tc.ie, tc.ieOk, "")

			if out, err = tc.ie.MarshalBinary(); err != nil {
				t.Error(err)
			}
			assert.Equal(t, out, tc.out, "")
		})
	}
}
