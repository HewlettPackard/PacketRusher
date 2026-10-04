// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestBackendValidationBeforeLoadingConfigOrReports(t *testing.T) {
	called := false
	app := parserApp(t, func(*cli.Context) error { called = true; return nil })
	err := app.Run([]string{"packetrusher", "--config", "/does/not/exist", "--tunnel-backend", "nonsense", "multi-ue", "-n", "2"})
	require.ErrorContains(t, err, "must be gtp5g, userspace, or ebpf")
	require.False(t, called)
	app = parserApp(t, func(c *cli.Context) error {
		called = true
		require.Equal(t, "userspace", c.String("tunnel-backend"))
		return nil
	})
	require.NoError(t, app.Run([]string{"packetrusher", "--tunnel-backend", "userspace", "multi-ue", "-n", "2", "--tunnel"}))
	require.True(t, called)
}

func TestEBPFUnsupportedProfileRejectedBeforeNetwork(t *testing.T) {
	for _, args := range [][]string{{"multi-ue", "-n", "1", "--tunnel", "--numPduSessions", "2"}} {
		app := newApp()
		err := app.Run(append([]string{"packetrusher", "--tunnel-backend", "ebpf", "--config", "../config/config.yml"}, args...))
		require.ErrorContains(t, err, "eBPF")
	}
}
