// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIPv6KernelTunnelRejectedBeforeNetworkOrCapture(t *testing.T) {
	contents, err := os.ReadFile("../config/config.yml")
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "ipv6.yml")
	require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(contents), "pdusessiontype: IPv4", "pdusessiontype: IPv6", 1)), 0600))
	for _, arguments := range [][]string{{"ue"}, {"multi-ue", "-n", "1", "--tunnel"}} {
		app := newApp()
		err := app.Run(append([]string{"packetrusher", "--config", path}, arguments...))
		require.ErrorContains(t, err, "requires ue.tunnelbackend: userspace")
	}
}
