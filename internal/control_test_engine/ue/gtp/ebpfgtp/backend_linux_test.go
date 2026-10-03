//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"errors"
	"github.com/cilium/ebpf"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActualVerifierLoadDeclaredCapabilities(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_RESTRICTED_CAPS") != "1" {
		t.Skip("requires the native runner's restricted capability profile")
	}
	status, err := os.ReadFile("/proc/self/status")
	require.NoError(t, err)
	var effective uint64
	var found bool
	for _, line := range strings.Split(string(status), "\n") {
		if value, ok := strings.CutPrefix(line, "CapEff:\t"); ok {
			effective, err = strconv.ParseUint(strings.TrimSpace(value), 16, 64)
			require.NoError(t, err)
			found = true
		}
	}
	require.True(t, found, "must inspect the actual process capabilities")
	require.NotZero(t, effective&(uint64(1)<<12), "CAP_NET_ADMIN required")
	require.NotZero(t, effective&(uint64(1)<<39), "CAP_BPF required")
	require.Zero(t, effective&(uint64(1)<<21), "must not rely on CAP_SYS_ADMIN")
	require.Zero(t, effective&(uint64(1)<<38), "must not rely on CAP_PERFMON")
	t.Logf("CapEff=%016x; BPF/NET_ADMIN present, SYS_ADMIN/PERFMON absent", effective)
	TestActualVerifierLoad(t)
}

func TestActualVerifierLoad(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires isolated privileged namespace; see docs/ebpf-backend.md")
	}
	r := NewRegistry()
	err := r.load()
	var verifier *ebpf.VerifierError
	if errors.As(err, &verifier) {
		t.Logf("%+v", verifier)
	}
	require.NoError(t, err)
	require.NotNil(t, r.collection.Programs["encap"])
	require.NotNil(t, r.collection.Programs["decap"])
	r.closeEmpty()
}
