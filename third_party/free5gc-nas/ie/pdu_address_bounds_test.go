// SPDX-License-Identifier: Apache-2.0
package ie

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPDUAddressOwnedExactLengthAndDecoderReuse(t *testing.T) {
	v6 := []byte{2, 0, 0, 0, 0, 0, 0, 0, 7}
	var address PDUAddr
	require.NoError(t, address.UnmarshalBinary(v6))
	v6[8] = 99
	require.Equal(t, byte(7), address.IPv6IfId[7], "decoder owns its value")
	require.NoError(t, address.UnmarshalBinary([]byte{1, 10, 0, 0, 2}))
	require.Nil(t, address.IPv6IfId, "decoder reuse clears earlier address family")
	for _, wire := range [][]byte{{2, 0, 0, 0, 0, 0, 0, 7}, {2, 0, 0, 0, 0, 0, 0, 0, 7, 0}, {3, 0, 0, 0, 0, 0, 0, 0, 7, 10, 0, 0}, {0x0a, 0, 0, 0, 0, 0, 0, 0, 7}} {
		require.Error(t, address.UnmarshalBinary(wire))
		require.Equal(t, []byte{10, 0, 0, 2}, address.IPv4, "failed parse preserves prior value")
	}
	for _, partial := range []*PDUAddr{{IPv4: []byte{1}}, {IPv6IfId: []byte{7}}, {IPv4: []byte{10, 0, 0, 2}, SMFIPv6LLA: []byte{1}}} {
		_, err := partial.MarshalBinary()
		require.Error(t, err)
	}
}
