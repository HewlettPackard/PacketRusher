// SPDX-License-Identifier: Apache-2.0
package ngap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLocalAssociationPortKeepsKernelAssignment(t *testing.T) {
	for _, offset := range []int32{0, 1, 100, 65536} {
		require.Zero(t, localAssociationPort(0, offset))
		require.Equal(t, uint16(9489+uint16(offset)), localAssociationPort(9489, offset))
	}
}
