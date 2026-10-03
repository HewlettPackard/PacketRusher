// SPDX-License-Identifier: Apache-2.0
package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControlClientValidationAvoidsConfigAndReports(t *testing.T) {
	for _, args := range [][]string{
		{"control", "--socket", "/absent", "--action", "xn-handover", "--ue", "1"},
		{"control", "--socket", "/absent", "--action", "idle", "--ue", "0"},
		{"control", "--socket", "/absent", "--timeout", "0"},
		{"control", "--socket", "/absent", "--action", "something"},
	} {
		report := filepath.Join(t.TempDir(), "results.json")
		app := newApp()
		app.Writer = io.Discard
		app.ErrWriter = io.Discard
		argv := append([]string{"packetrusher", "--config", "/absent/config", "--report-json", report}, args...)
		err := app.Run(argv)
		if err == nil || strings.Contains(err.Error(), "config") {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if _, err := os.Stat(report); !os.IsNotExist(err) {
			t.Fatal("control client reserved a report")
		}
	}
}
