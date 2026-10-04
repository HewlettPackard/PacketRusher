/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package main

import (
	"encoding/json"
	"io"
	"my5G-RANTester/internal/analytics"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// The report is written when the command ends, even with an error, and replaces
// the report of a previous run.
func TestReportIsWrittenAfterActionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	require.NoError(t, os.WriteFile(path, []byte("previous run"), 0o644))
	app := newApp()
	app.Writer, app.ErrWriter = io.Discard, io.Discard
	require.Error(t, app.Run([]string{"packetrusher", "--report-json", path, "control", "--socket", filepath.Join(dir, "missing")}))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var report struct{ Procedures []analytics.Result }
	require.NoError(t, json.Unmarshal(data, &report))
	require.Len(t, report.Procedures, 2)
	require.Equal(t, "registration", report.Procedures[0].Procedure)
}
