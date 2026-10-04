/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package control

import (
	"my5G-RANTester/internal/control_test_engine/procedures"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestValidate(t *testing.T) {
	for _, test := range []struct {
		request Request
		valid   bool
	}{
		{Request{Action: "inspect"}, true},
		{Request{UeId: 2, Action: "inspect"}, true},
		{Request{UeId: 1, Action: "wait", TimeoutMs: 45000}, true},
		{Request{UeId: 1, Action: "deregister"}, true},
		{Request{UeId: 1, Action: "ng-handover", Target: "000009"}, true},
		{Request{UeId: 1}, false},
		{Request{UeId: 1, Action: "terminate"}, false},
		{Request{Action: "wait"}, false},
		{Request{UeId: -1, Action: "inspect"}, false},
		{Request{UeId: 1, Action: "xn-handover"}, false},
		{Request{UeId: 1, Action: "idle", Target: "000009"}, false},
		{Request{UeId: 1, Action: "idle", TimeoutMs: -1}, false},
	} {
		if err := test.request.Validate(); (err == nil) != test.valid {
			t.Errorf("Validate(%+v) = %v, want valid: %t", test.request, err, test.valid)
		}
	}
}

func TestListen(t *testing.T) {
	// A file which is not a socket is kept.
	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, nil, 0600))
	_, err := Listen(file, nil, 2)
	require.Error(t, err)
	require.FileExists(t, file)

	// A socket left behind by a crash is replaced, unlike one in use.
	socket := filepath.Join(t.TempDir(), "control.sock")
	stale, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket})
	require.NoError(t, err)
	stale.SetUnlinkOnClose(false)
	require.NoError(t, stale.Close())
	server, err := Listen(socket, nil, 2)
	require.NoError(t, err)
	_, err = Listen(socket, nil, 2)
	require.ErrorContains(t, err, "already in use")
	info, err := os.Stat(socket)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// A UE which is not started yet can only be inspected.
	response, err := Call(socket, Request{Action: "inspect"})
	require.NoError(t, err)
	require.Equal(t, []procedures.UeStatus{{UeId: 1, State: "starting", PduSessions: []int{}}, {UeId: 2, State: "starting", PduSessions: []int{}}}, response.Ues)
	_, err = Call(socket, Request{UeId: 2, Action: "idle"})
	require.ErrorContains(t, err, "not started")
	_, err = Call(socket, Request{UeId: 3, Action: "inspect"})
	require.ErrorContains(t, err, "unknown UE 3")
	_, err = Call(socket, Request{UeId: 1, Action: "terminate"})
	require.ErrorContains(t, err, "unknown action")

	require.NoError(t, server.Close())
	require.NoFileExists(t, socket)
}
