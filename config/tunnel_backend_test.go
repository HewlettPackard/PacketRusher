// SPDX-License-Identifier: Apache-2.0
package config

import (
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

func TestTunnelBackendDefaultsAndValidation(t *testing.T) {
	for _, name := range []string{"", "auto", "gtp5g", "userspace", "ebpf"} {
		_, err := ParseTunnelBackend(name)
		require.NoError(t, err)
	}
	backend, err := ParseTunnelBackend("")
	require.NoError(t, err)
	require.Equal(t, TunnelBackendAuto, backend)
	_, err = ParseTunnelBackend("kernel")
	require.Error(t, err)
}

func TestEBPFProfileChecksPortAndAcceptsDualStackJumboMTU(t *testing.T) {
	cfg := Config{}
	cfg.Ue.TunnelBackend = TunnelBackendEBPF
	require.ErrorContains(t, cfg.ValidateTunnel(true), "dataif.port: 2152")
	cfg.GNodeB.DataIF.AddrPort = netip.MustParseAddrPort("192.0.2.1:2152")
	require.NoError(t, cfg.ValidateTunnel(true))
	cfg.Ue.TunnelMTU = 1457
	require.NoError(t, cfg.ValidateTunnel(true), "TCX is no longer restricted to Ethernet MTU 1500")
	cfg.Ue.PDUSessionType = PDUSessionType(3)
	require.NoError(t, cfg.ValidateTunnel(true))
	cfg.Ue.TunnelMTU = 1279
	require.ErrorContains(t, cfg.ValidateTunnel(true), "1280")
	require.NoError(t, cfg.ValidateTunnel(false))
}

func TestOmittedBackendDefaultsToAutoWithoutChangingExplicitChoices(t *testing.T) {
	for _, value := range []string{"", "auto", "ebpf", "userspace", "gtp5g"} {
		got, err := ParseTunnelBackend(value)
		require.NoError(t, err)
		if value == "" {
			require.Equal(t, TunnelBackendAuto, got)
		} else {
			require.Equal(t, TunnelBackend(value), got)
		}
	}
	cfg := Config{Ue: Ue{PDUSessionType: PDUSessionType(3)}}
	require.ErrorContains(t, cfg.ValidateTunnel(true), "2152")
	cfg.GNodeB.DataIF.AddrPort = netip.MustParseAddrPort("192.0.2.1:2152")
	require.NoError(t, cfg.ValidateTunnel(true))
}

func TestPortableBackendsRejectUnsupportedCustomN3Port(t *testing.T) {
	for _, backend := range []TunnelBackend{TunnelBackendAuto, TunnelBackendEBPF, TunnelBackendUserspace} {
		cfg := Config{Ue: Ue{TunnelBackend: backend}}
		cfg.GNodeB.DataIF.AddrPort = netip.MustParseAddrPort("192.0.2.1:9999")
		require.ErrorContains(t, cfg.ValidateTunnel(true), "both eBPF and userspace bind UDP/2152")
		require.NoError(t, cfg.ValidateTunnel(false), "control-plane-only profiles do not bind GTP")
	}
}
