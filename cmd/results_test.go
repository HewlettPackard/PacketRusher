// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"my5G-RANTester/internal/analytics"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestReportsWrittenAfterActionError(t *testing.T) {
	dir := t.TempDir()
	jsonPath, csvPath := filepath.Join(dir, "result.json"), filepath.Join(dir, "result.csv")
	actionErr := errors.New("scenario stopped")
	app := newApp()
	app.Writer, app.ErrWriter = io.Discard, io.Discard
	app.Command("multi-ue").Action = func(*cli.Context) error {
		r := analytics.Current()
		require.NotNil(t, r)
		r.Begin(1, 0, analytics.Registration)
		r.Finish(1, 0, analytics.Registration, analytics.Success)
		r.Begin(2, 0, analytics.Registration)
		return actionErr
	}
	require.ErrorIs(t, app.Run([]string{"packetrusher", "--report-json", jsonPath, "--report-csv", csvPath, "multi-ue", "-n", "1"}), actionErr)
	data, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	var report analytics.Report
	require.NoError(t, json.Unmarshal(data, &report))
	require.NotNil(t, report.EndedAt)
	for _, a := range report.Procedures {
		if a.Procedure == analytics.Registration {
			require.Equal(t, uint64(1), a.Success)
			require.Equal(t, uint64(1), a.Cancelled)
		}
	}
	require.Nil(t, analytics.Current())
	csvData, err := os.ReadFile(csvPath)
	require.NoError(t, err)
	require.Contains(t, string(csvData), "registration,2,1,0,1,0,1,")
}

func TestExistingReportIsPreserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.json")
	require.NoError(t, os.WriteFile(path, []byte("existing result"), 0644))
	before, after := resultsHooks()
	app := &cli.App{Flags: []cli.Flag{&cli.PathFlag{Name: "report-json"}}, Before: before, After: after, Action: func(*cli.Context) error { t.Fatal("must fail before scenario"); return nil }}
	require.ErrorContains(t, app.Run([]string{"packetrusher", "--report-json", path}), "file exists")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "existing result", string(data))
}

func TestRejectedCommandLeavesReportPathsAvailable(t *testing.T) {
	for _, rejected := range [][]string{
		{"-n=invalid"},
		{"--tunnel", "true", "-n", "2"},
		{"-n", "0"},
		{"-n", "2", "--numPduSessions=0"},
		{"-n", "2", "--timeBetweenRegistration=-1"},
	} {
		t.Run(fmt.Sprint(rejected), func(t *testing.T) {
			dir := t.TempDir()
			jsonPath, csvPath := filepath.Join(dir, "result.json"), filepath.Join(dir, "result.csv")
			called := false
			app := newApp()
			app.Writer, app.ErrWriter = io.Discard, io.Discard
			app.Command("multi-ue").Action = func(*cli.Context) error { called = true; return nil }
			args := append([]string{"packetrusher", "--report-json", jsonPath, "--report-csv", csvPath, "multi-ue"}, rejected...)
			require.Error(t, app.Run(args))
			require.False(t, called, "rejected commands must not start a scenario")
			require.Nil(t, analytics.Current())
			for _, path := range []string{jsonPath, csvPath} {
				_, err := os.Stat(path)
				require.ErrorIs(t, err, os.ErrNotExist)
			}

			// Correct the syntax and reuse the same output paths with the real parser.
			next := newApp()
			next.Writer, next.ErrWriter = io.Discard, io.Discard
			next.Command("multi-ue").Action = func(*cli.Context) error {
				called = true
				recorder := analytics.Current()
				require.NotNil(t, recorder)
				recorder.Begin(1, 0, analytics.Registration)
				recorder.Finish(1, 0, analytics.Registration, analytics.Success)
				return nil
			}
			require.NoError(t, next.Run([]string{"packetrusher", "--report-json", jsonPath, "--report-csv", csvPath, "multi-ue", "-n", "2", "--tunnel"}))
			require.True(t, called)
			require.Nil(t, analytics.Current())
			data, err := os.ReadFile(jsonPath)
			require.NoError(t, err)
			var report analytics.Report
			require.NoError(t, json.Unmarshal(data, &report))
			require.NotNil(t, report.EndedAt)
			for _, procedure := range report.Procedures {
				if procedure.Procedure == analytics.Registration {
					require.Equal(t, uint64(1), procedure.Started)
					require.Equal(t, uint64(1), procedure.Success)
				}
			}
			csvData, err := os.ReadFile(csvPath)
			require.NoError(t, err)
			require.Contains(t, string(csvData), "registration,1,1,0,0,0,1,")
		})
	}
}
