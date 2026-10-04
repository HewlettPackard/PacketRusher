// SPDX-License-Identifier: Apache-2.0
package config

import (
	"github.com/stretchr/testify/require"
	"net/netip"
	"testing"
)

func TestTunnelBackendDefaultsAndValidation(t *testing.T) {
	for _, name := range []string{"", "gtp5g", "userspace", "ebpf"} {
		_, err := ParseTunnelBackend(name)
		require.NoError(t, err)
	}
	backend, err := ParseTunnelBackend("")
	require.NoError(t, err)
	require.Equal(t, TunnelBackendKernel, backend)
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
