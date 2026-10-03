// SPDX-License-Identifier: Apache-2.0
package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTunnelBackendDefaultsAndValidation(t *testing.T) {
	for _, name := range []string{"", "gtp5g", "userspace"} {
		_, err := ParseTunnelBackend(name)
		require.NoError(t, err)
	}
	backend, err := ParseTunnelBackend("")
	require.NoError(t, err)
	require.Equal(t, TunnelBackendKernel, backend)
	_, err = ParseTunnelBackend("kernel")
	require.Error(t, err)
}
