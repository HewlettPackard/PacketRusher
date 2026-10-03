// SPDX-License-Identifier: Apache-2.0
package context

import (
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

func TestMockIPv6AllocationUsesUniquePrefixAndDoesNotWrap(t *testing.T) {
	session := new(SessionContext)
	session.NewSessionContext()
	first, err := session.GetUnallocatedIPv6()
	require.NoError(t, err)
	second, err := session.GetUnallocatedIPv6()
	require.NoError(t, err)
	one, _ := netip.AddrFromSlice(first)
	two, _ := netip.AddrFromSlice(second)
	require.NotEqual(t, netip.PrefixFrom(one, 64), netip.PrefixFrom(two, 64))
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 1}, []byte(first[8:]))
	session.lastAllocatedIPv6 = 1<<32 - 1
	_, err = session.GetUnallocatedIPv6()
	require.ErrorContains(t, err, "pool exhausted")
}
