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
	require.ErrorContains(t, err, "must be gtp5g or userspace")
	require.False(t, called)
	app = parserApp(t, func(c *cli.Context) error {
		called = true
		require.Equal(t, "userspace", c.String("tunnel-backend"))
		return nil
	})
	require.NoError(t, app.Run([]string{"packetrusher", "--tunnel-backend", "userspace", "multi-ue", "-n", "2", "--tunnel"}))
	require.True(t, called)
}
