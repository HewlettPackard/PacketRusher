/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package config

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTunnelBackend(t *testing.T) {
	saved := slices.Clone(tunnelBackends)
	t.Cleanup(func() { copy(tunnelBackends, saved) })
	reason := func(available bool) func() error {
		return func() error {
			if available {
				return nil
			}
			return errors.New("missing")
		}
	}
	for _, tc := range []struct {
		name        string
		session     PDUSessionType
		ebpf, gtp5g bool
		want        TunnelBackend
		err         string
	}{
		{"auto", "", true, true, TunnelBackendEBPF, ""},
		{"auto", "", false, true, TunnelBackendGtp5g, ""},
		{"auto", "", false, false, TunnelBackendUserspace, ""},
		{"auto", PDUSessionIPv6, false, true, TunnelBackendUserspace, ""},
		{"ebpf", "", false, true, "", "not available: missing"},
		{"gtp5g", "", true, false, "", "not available: missing"},
		{"gtp5g", PDUSessionIPv4v6, true, true, TunnelBackendGtp5g, ""},
		{"gtp5g", PDUSessionIPv6, true, true, "", "cannot carry IPv6"},
		{"userspace", "", true, true, TunnelBackendUserspace, ""},
		{"kernel", "", true, true, "", "unknown tunnel backend"},
	} {
		tunnelBackends[0].unavailable, tunnelBackends[1].unavailable = reason(tc.ebpf), reason(tc.gtp5g)
		backend, err := ResolveTunnelBackend(tc.name, tc.session)
		require.Equal(t, tc.want, backend, "%+v", tc)
		if tc.err == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.err)
		}
	}
}
