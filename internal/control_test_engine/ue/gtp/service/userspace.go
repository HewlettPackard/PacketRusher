/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"fmt"
	"net/netip"

	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

// userspaceTunnel is the userspace datapath: a TUN device whose packets this
// process carries over GTP-U itself.
type userspaceTunnel struct{ *userspace.Session }

func userspaceConfig(pdu *gnbContext.GnbPDUSession, ueIP string) (userspace.Config, error) {
	upf, err := netip.ParseAddr(pdu.GetUpfIp())
	if err != nil {
		return userspace.Config{}, fmt.Errorf("UPF address: %w", err)
	}
	ue, err := netip.ParseAddr(ueIP)
	if err != nil {
		return userspace.Config{}, fmt.Errorf("UE address: %w", err)
	}
	qfi := pdu.GetQosId()
	if qfi < 0 || qfi > 63 {
		return userspace.Config{}, fmt.Errorf("QFI %d is not between 0 and 63", qfi)
	}
	return userspace.Config{UPF: upf, UplinkTEID: pdu.GetTeidUplink(), DownlinkTEID: pdu.GetTeidDownlink(), QFI: uint8(qfi), UE: ue}, nil
}

func startUserspace(name string, pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) (datapath, error) {
	config, err := userspaceConfig(pdu, ueIP)
	if err != nil {
		return nil, err
	}
	session, err := userspace.Open(name, ip, config)
	if err != nil {
		return nil, err
	}
	return userspaceTunnel{session}, nil
}

func (t userspaceTunnel) refresh(pdu *gnbContext.GnbPDUSession, ueIP string, _ netip.Addr) error {
	config, err := userspaceConfig(pdu, ueIP)
	if err == nil {
		t.Update(config)
	}
	return err
}

func (t userspaceTunnel) close() { t.Close() }
