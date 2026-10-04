// SPDX-License-Identifier: Apache-2.0
package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/free5gc/nas/ie"
)

// PDUSessionType selects the requested inner IP family; zero preserves IPv4.
type PDUSessionType uint8

func (t PDUSessionType) NASValue() uint8 {
	if t == 0 {
		return ie.PDUSessType_IPv4
	}
	return uint8(t)
}

// ValidateTunnel checks requested families before starting telecom sockets.
// IPv6 control-plane scenarios remain available when tunnels are disabled.
func (c Config) ValidateTunnel(enabled bool) error {
	if !enabled {
		return nil
	}
	backend, err := ParseTunnelBackend(string(c.Ue.TunnelBackend))
	if err != nil {
		return err
	}
	if backend != TunnelBackendKernel && c.GNodeB.DataIF.Port() != 2152 {
		return fmt.Errorf("%s tunnels require gnodeb.dataif.port: 2152; both eBPF and userspace bind UDP/2152", backend)
	}
	if c.Ue.PDUSessionType.NASValue() == ie.PDUSessType_IPv4 {
		return nil
	}
	if backend != TunnelBackendUserspace && backend != TunnelBackendEBPF && backend != TunnelBackendAuto {
		return fmt.Errorf("IPv6/IPv4v6 user traffic requires ue.tunnelbackend: userspace, ebpf, or auto")
	}
	if c.Ue.TunnelMTU != 0 && c.Ue.TunnelMTU < 1280 {
		return fmt.Errorf("IPv6/IPv4v6 requires ue.tunnelmtu of at least 1280, or 0 for automatic selection")
	}
	return nil
}

func (t *PDUSessionType) UnmarshalYAML(_ context.Context, unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	switch strings.ToLower(value) {
	case "", "ipv4":
		*t = PDUSessionType(ie.PDUSessType_IPv4)
	case "ipv6":
		*t = PDUSessionType(ie.PDUSessType_IPv6)
	case "ipv4v6":
		*t = PDUSessionType(ie.PDUSessType_IPv4v6)
	default:
		return fmt.Errorf("ue.pdusessiontype %q must be IPv4, IPv6, or IPv4v6", value)
	}
	return nil
}
