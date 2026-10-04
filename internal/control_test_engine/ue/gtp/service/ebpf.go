/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"net/netip"

	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
)

// ebpfTunnel is the userspace datapath with its packets short-cut in the kernel.
type ebpfTunnel struct {
	userspaceTunnel
	name string
}

func startEBPF(name string, pdu *gnbContext.GnbPDUSession, ue *context.UEPDUSession, ip netip.Addr) (datapath, error) {
	tunnel, err := startUserspace(name, pdu, ue, ip)
	if err != nil {
		return nil, err
	}
	t := ebpfTunnel{tunnel.(userspaceTunnel), name}
	config, _ := userspaceConfig(pdu, ue)
	if err := ebpfgtp.Attach(name, ip, config); err != nil {
		t.close()
		return nil, err
	}
	return t, nil
}

func (t ebpfTunnel) refresh(pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) error {
	if err := t.userspaceTunnel.refresh(pdu, ueIP, ip); err != nil {
		return err
	}
	config, _ := userspaceConfig(pdu, t.session)
	return ebpfgtp.Attach(t.name, ip, config)
}

func (t ebpfTunnel) close() {
	ebpfgtp.Detach(t.name)
	t.userspaceTunnel.close()
}
