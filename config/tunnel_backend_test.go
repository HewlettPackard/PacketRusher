/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTunnelBackend(t *testing.T) {
	gtp5g := tunnelBackends[0].available
	t.Cleanup(func() { tunnelBackends[0].available = gtp5g })
	for _, tc := range []struct {
		name    string
		session PDUSessionType
		loaded  bool
		want    TunnelBackend
		err     string
	}{
		{"auto", "", true, TunnelBackendGtp5g, ""},
		{"auto", "", false, TunnelBackendUserspace, ""},
		{"gtp5g", PDUSessionIPv4v6, true, TunnelBackendGtp5g, ""},
		{"gtp5g", "", false, "", "not available"},
		{"userspace", "", true, TunnelBackendUserspace, ""},
		{"kernel", "", true, "", "unknown tunnel backend"},
		{"auto", PDUSessionIPv6, true, TunnelBackendUserspace, ""},
		{"gtp5g", PDUSessionIPv6, true, "", "cannot carry IPv6"},
	} {
		tunnelBackends[0].available = func() bool { return tc.loaded }
		backend, err := ResolveTunnelBackend(tc.name, tc.session)
		require.Equal(t, tc.want, backend, "%s, %s, gtp5g loaded: %v", tc.name, tc.session, tc.loaded)
		if tc.err == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.err)
		}
	}
}
