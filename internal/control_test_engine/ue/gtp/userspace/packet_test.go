/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package userspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	ip := []byte{0x45, 1, 2, 3}
	plain := []byte{0x30, 0xff, 0, 4, 0, 0, 0, 42}
	container := []byte{0x34, 0xff, 0, 12, 0, 0, 0, 42, 0, 0, 0, 0x85, 1, 0x10, 9, 0}
	buffer := append(make([]byte, Headroom), ip...)
	require.Equal(t, append(plain, ip...), Encode(buffer, 42, 0))
	require.Equal(t, append(container, ip...), Encode(buffer, 42, 9))
	require.Zero(t, testing.AllocsPerRun(10, func() { Encode(buffer, 42, 9) }))

	for name, tc := range map[string]struct {
		header []byte
		valid  bool
	}{
		"no optional field":         {plain, true},
		"PDU Session Container":     {container, true},
		"two extension headers":     {[]byte{0x34, 0xff, 0, 16, 0, 0, 0, 42, 0, 0, 0, 0x40, 1, 8, 0x68, 0x85, 1, 0, 9, 0}, true},
		"spare bit set":             {[]byte{0x38, 0xff, 0, 4, 0, 0, 0, 42}, true},
		"next extension without E":  {[]byte{0x32, 0xff, 0, 8, 0, 0, 0, 42, 0, 7, 0, 0x85}, true},
		"padding after the G-PDU":   {[]byte{0x30, 0xff, 0, 4, 0, 0, 0, 42, 0x45, 1, 2, 3}, true},
		"GTPv2":                     {[]byte{0x50, 0xff, 0, 4, 0, 0, 0, 42}, false},
		"not a G-PDU":               {[]byte{0x30, 26, 0, 4, 0, 0, 0, 42}, false},
		"shorter than its length":   {[]byte{0x30, 0xff, 0, 5, 0, 0, 0, 42}, false},
		"optional fields cut short": {[]byte{0x32, 0xff, 0, 2, 0, 0, 0, 42}, false},
		"empty extension header":    {[]byte{0x34, 0xff, 0, 12, 0, 0, 0, 42, 0, 0, 0, 0x85, 0, 0x10, 9, 0}, false},
		"extension past the end":    {[]byte{0x34, 0xff, 0, 12, 0, 0, 0, 42, 0, 0, 0, 0x85, 3, 0x10, 9, 0}, false},
	} {
		teid, packet, err := Decode(append(append([]byte{}, tc.header...), ip...))
		if !tc.valid {
			require.Error(t, err, name)
			continue
		}
		require.NoError(t, err, name)
		require.Equal(t, uint32(42), teid, name)
		require.Equal(t, ip, packet, name)
	}
	_, _, err := Decode(plain[:7])
	require.Error(t, err)

	request := []byte{0x32, 1, 0, 4, 0, 0, 0, 0, 0x12, 0x34, 0, 0}
	require.Equal(t, []byte{0x32, 2, 0, 6, 0, 0, 0, 0, 0x12, 0x34, 0, 0, 14, 0}, echoResponse(request))
	require.Nil(t, echoResponse(append(plain, ip...)))
}
