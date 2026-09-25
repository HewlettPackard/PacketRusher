/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package mm_5gs

import (
	"bytes"
	"testing"

	"github.com/free5gc/nas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeImeisv reads the IMEISV back out of an encoded Security Mode Complete and applies the
// TS 24.501 9.11.3.4 rules strictly: every nibble is a BCD digit except the last, which must
// be the "1111" end mark. The free5gc decoder drops the last nibble whatever it holds, so it
// cannot tell a missing end mark from a present one; this check can.
func decodeImeisv(t *testing.T, pdu []byte) string {
	t.Helper()

	m := nas.NewMessage()
	require.NoError(t, m.GmmMessageDecode(&pdu))
	require.NotNil(t, m.SecurityModeComplete.IMEISV, "Security Mode Complete should carry an IMEISV")

	imeisv := m.SecurityModeComplete.IMEISV
	require.Equal(t, uint16(9), imeisv.GetLen())
	assert.Equal(t, uint8(5), imeisv.GetTypeOfIdentity(), "type of identity should be IMEISV")
	assert.Equal(t, uint8(0), imeisv.GetOddEvenIdic(), "an IMEISV has an even number of digits")

	var digits bytes.Buffer
	nibbles := []uint8{imeisv.Octet[0] >> 4}
	for _, octet := range imeisv.Octet[1:imeisv.GetLen()] {
		nibbles = append(nibbles, octet&0x0f, octet>>4)
	}
	last := len(nibbles) - 1
	assert.Equal(t, uint8(0x0f), nibbles[last], "bits 5 to 8 of the last octet should be the 1111 end mark")
	for _, n := range nibbles[:last] {
		require.LessOrEqual(t, n, uint8(9), "every IMEISV nibble but the end mark should be a BCD digit")
		digits.WriteByte('0' + n)
	}
	return digits.String()
}

func TestSecurityModeCompleteImeisvDecodes(t *testing.T) {
	imeisv := decodeImeisv(t, getSecurityModeComplete(nil, imeisvFromMsin("0000000120")))

	assert.Equal(t, "0000000000012001", imeisv)
}

// A constant IMEISV would make every UE in a scale run indistinguishable by PEI.
func TestSecurityModeCompleteImeisvIsPerUe(t *testing.T) {
	first := decodeImeisv(t, getSecurityModeComplete(nil, imeisvFromMsin("0000000120")))
	second := decodeImeisv(t, getSecurityModeComplete(nil, imeisvFromMsin("0000000121")))

	assert.NotEqual(t, first, second)
}

func TestImeisvFromMsin(t *testing.T) {
	assert.Equal(t, "0000012345678901", imeisvFromMsin("123456789"), "a 9-digit MSIN is left-padded")
	assert.Equal(t, "0000000000012001", imeisvFromMsin("0000000120"))
	assert.Len(t, imeisvFromMsin("123456789012345678"), 16, "an over-long input still yields 16 digits")
}
