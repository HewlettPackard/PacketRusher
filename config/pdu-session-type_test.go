// SPDX-License-Identifier: Apache-2.0
package config

import (
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

func TestPDUSessionTypeConfig(t *testing.T) {
	for _, tc := range []struct {
		value    string
		expected uint8
	}{{"", 1}, {"IPv4", 1}, {"IPv6", 2}, {"IPv4v6", 3}, {"ipv4V6", 3}} {
		t.Run(tc.value, func(t *testing.T) {
			var cfg Config
			document := "ue: {}"
			if tc.value != "" {
				document = "ue:\n  pdusessiontype: " + tc.value
			}
			require.NoError(t, yaml.UnmarshalWithOptions([]byte(document), &cfg, yaml.Strict()))
			require.Equal(t, tc.expected, cfg.Ue.PDUSessionType.NASValue())
		})
	}
	for _, invalid := range []string{"IPv5", "Ethernet", "42"} {
		var cfg Config
		require.Error(t, yaml.UnmarshalWithOptions([]byte("ue:\n  pdusessiontype: "+invalid), &cfg, yaml.Strict()))
	}
}

func TestIPv6TunnelPreflight(t *testing.T) {
	cfg := Config{Ue: Ue{PDUSessionType: PDUSessionType(2), TunnelBackend: TunnelBackendKernel}}
	cfg.GNodeB.DataIF.AddrPort = netip.MustParseAddrPort("192.0.2.1:2152")
	require.NoError(t, cfg.ValidateTunnel(false))
	require.ErrorContains(t, cfg.ValidateTunnel(true), "userspace")
	cfg.Ue.TunnelBackend = TunnelBackendUserspace
	require.NoError(t, cfg.ValidateTunnel(true))
	cfg.Ue.TunnelMTU = 1279
	require.ErrorContains(t, cfg.ValidateTunnel(true), "1280")
	cfg.Ue.TunnelMTU = 1280
	require.NoError(t, cfg.ValidateTunnel(true))
	cfg.Ue.PDUSessionType = PDUSessionType(1)
	cfg.Ue.TunnelBackend = TunnelBackendKernel
	cfg.Ue.TunnelMTU = 1400
	require.NoError(t, cfg.ValidateTunnel(true))
}
