package ie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrityProtectionMaxDataRateUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		ie    IntegrityProtectionMaxDataRate
		input []byte
		ieOk  IntegrityProtectionMaxDataRate
		out   []byte
	}{
		{
			name:  "",
			ie:    IntegrityProtectionMaxDataRate{},
			input: []byte{0x2E, 0x14},
			ieOk: IntegrityProtectionMaxDataRate{
				Uplink:   0x2E,
				Downlink: 0x14,
			},
			out: []byte{0x2E, 0x14},
		},
		{
			name:  "",
			ie:    IntegrityProtectionMaxDataRate{},
			input: []byte{0x8F, 0xA8},
			ieOk: IntegrityProtectionMaxDataRate{
				Uplink:   0x8F,
				Downlink: 0xA8,
			},
			out: []byte{0x8F, 0xA8},
		},
		{
			name:  "",
			ie:    IntegrityProtectionMaxDataRate{},
			input: []byte{0xA2, 0x99},
			ieOk: IntegrityProtectionMaxDataRate{
				Uplink:   0xA2,
				Downlink: 0x99,
			},
			out: []byte{0xA2, 0x99},
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
