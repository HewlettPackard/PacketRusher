// SPDX-License-Identifier: Apache-2.0
package scenario

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestControlSocketOwnershipAndJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "control.sock")
	server, err := Listen(path, NewRegistry(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("socket permissions %v", info.Mode().Perm())
	}
	if _, err := Listen(path, NewRegistry(nil)); err == nil {
		t.Fatal("replaced a live control socket")
	}
	response, err := Call(context.Background(), path, Request{Action: "inspect"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.UEs) != 0 {
		t.Fatal(response)
	}
	if _, err := Call(context.Background(), path, Request{Action: "idle", UE: 1}); err == nil || !strings.Contains(err.Error(), "unknown UE") {
		t.Fatalf("missing useful error: %v", err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("socket left behind: %v", err)
	}
	if err := os.WriteFile(path, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	server.Close()
	if content, err := os.ReadFile(path); err != nil || string(content) != "keep me" {
		t.Fatalf("cleanup removed unrelated replacement: %v", err)
	}
}

func TestListenKeepsExistingFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "important")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(path, NewRegistry(nil)); err == nil {
		t.Fatal("overwrote an existing file")
	}
	if got, _ := os.ReadFile(path); string(got) != "keep" {
		t.Fatal("existing content changed")
	}
}

func TestScenarioValidationBeforeConnecting(t *testing.T) {
	for _, input := range []string{
		`{"steps":[]}`, `{"steps":[{"ue":1,"action":"idle","extra":true}]}`,
		`{"steps":[{"ue":1,"action":"ng-handover"}]}`, `{"steps":[{"ue":0,"action":"deregister"}]}`,
		`{"steps":[{"ue":1,"action":"idle","timeout_ms":120001}]}`, `{"steps":[{"ue":1,"action":"idle"}]} {}`,
	} {
		if _, err := ReadPlan(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
	plan, err := ReadPlan(strings.NewReader(`{"steps":[{"ue":1,"action":"wait"},{"ue":1,"action":"xn-handover","target":"000009"}]}`))
	if err != nil || len(plan.Steps) != 2 {
		t.Fatalf("valid plan rejected: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := plan.Run(ctx, "/absent/socket", os.Stdout); err == nil || time.Since(start) > time.Second {
		t.Fatalf("cancelled scenario did not stop: %v", err)
	}
}
