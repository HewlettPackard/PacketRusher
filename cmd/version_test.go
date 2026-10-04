/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionNeedsNoConfiguration(t *testing.T) {
	var output bytes.Buffer
	app := newApp()
	app.Writer = &output
	require.NoError(t, app.Run([]string{"packetrusher", "--config", "/missing/config.yml", "--version"}))
	require.Contains(t, output.String(), "packetrusher version")
}
