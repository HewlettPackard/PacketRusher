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
	require.ErrorContains(t, err, "must be auto, gtp5g, userspace, or ebpf")
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

func TestAutoAllowsMultiplePDUNegotiationBeforeNetwork(t *testing.T) {
	// Omitting -n makes the production action return after configuration/preflight.
	// Auto retains userspace's multi-PDU negotiation, with traffic only on PDU 1.
	for _, backend := range []string{"auto", "userspace"} {
		require.NoError(t, newApp().Run([]string{"packetrusher", "--tunnel-backend", backend, "--config", "../config/config.yml", "multi-ue", "--tunnel", "--numPduSessions", "2"}))
	}
}

func TestCLIDefaultIsAutoAndExplicitBackendRemainsSelectable(t *testing.T) {
	for _, value := range []string{"", "auto", "ebpf", "userspace", "gtp5g"} {
		app := parserApp(t, func(c *cli.Context) error {
			if value == "" {
				require.Equal(t, "auto", c.String("tunnel-backend"))
				require.False(t, c.IsSet("tunnel-backend"))
			} else {
				require.Equal(t, value, c.String("tunnel-backend"))
				require.True(t, c.IsSet("tunnel-backend"))
			}
			return nil
		})
		args := []string{"packetrusher"}
		if value != "" {
			args = append(args, "--tunnel-backend", value)
		}
		require.NoError(t, app.Run(append(args, "multi-ue", "-n", "1", "--tunnel")))
	}
}
