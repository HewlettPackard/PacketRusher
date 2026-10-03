package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecProtectedHdrUnmarshalBinary(t *testing.T) {
	tcs := []struct {
		name      string
		input     []byte
		expectHdr *SecProtectedHdr
	}{
		{
			name:      "invalid len: 0",
			input:     []byte{},
			expectHdr: nil,
		},
		{
			name:      "invalid len: 6",
			input:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectHdr: nil,
		},
		{
			name:      "wrong epd: Epd5GSSessMgmtMsg",
			input:     []byte{0x2e, 0x02, 0xaa, 0xbb, 0xcc, 0xdd, 0xee},
			expectHdr: nil,
		},
		{
			name:  "valid input",
			input: []byte{0x7e, 0x02, 0xaa, 0xbb, 0xcc, 0xdd, 0xee},
			expectHdr: &SecProtectedHdr{
				SecHdrType:     SecHdrTypeIntegrityProtectedAndCiphered,
				MAC:            []byte{0xaa, 0xbb, 0xcc, 0xdd},
				SequenceNumber: 0xee,
			},
		},
	}
	for i := range tcs {
		tc := tcs[i]
		t.Run(tc.name, func(t *testing.T) {
			secHdr := new(SecProtectedHdr)
			err := secHdr.UnmarshalBinary(tc.input)
			if tc.expectHdr == nil {
				require.Error(t, err)
			} else {
				require.Equal(t, *tc.expectHdr, *secHdr)
			}
		})
	}
}
