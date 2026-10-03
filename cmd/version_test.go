// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"my5G-RANTester/internal/buildinfo"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionCommandDoesNotLoadConfigOrInitializeReports(t *testing.T) {
	for _, command := range [][]string{{"version", "--json"}, {"--version"}} {
		app := newApp()
		var output bytes.Buffer
		app.Writer = &output
		path := filepath.Join(t.TempDir(), "should-not-exist.json")
		args := append([]string{"packetrusher", "--config", "/missing/config.yml", "--report-json", path, "--metrics-addr", "bad-address"}, command...)
		if err := app.Run(args); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("version command reserved report output: %v", err)
		}
		if len(command) > 1 && command[1] == "--json" {
			var info buildinfo.Info
			if err := json.Unmarshal(output.Bytes(), &info); err != nil {
				t.Fatal(err)
			}
			if info.Version == "" || info.Revision == "" || info.GoVersion == "" {
				t.Fatalf("incomplete build info: %+v", info)
			}
		}
	}
}
