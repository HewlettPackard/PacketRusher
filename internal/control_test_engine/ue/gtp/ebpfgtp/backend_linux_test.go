//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"errors"
	"github.com/cilium/ebpf"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

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
