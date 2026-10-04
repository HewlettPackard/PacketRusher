/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"fmt"
	"net/netip"

	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

// userspaceTunnel is the userspace datapath: a TUN device whose packets this
// process carries over GTP-U itself, for the addresses of the UE's PDU session.
type userspaceTunnel struct {
	*userspace.Session
	session *context.UEPDUSession
}

func userspaceConfig(pdu *gnbContext.GnbPDUSession, session *context.UEPDUSession) (userspace.Config, error) {
	upf, err := netip.ParseAddr(pdu.GetUpfIp())
	if err != nil {
		return userspace.Config{}, fmt.Errorf("UPF address: %w", err)
	}
	ue, err := netip.ParseAddr(session.GetIp())
	if err != nil && session.GetIp() != "" {
		return userspace.Config{}, fmt.Errorf("UE address: %w", err)
	}
	var prefix netip.Prefix
	if ipv6 := session.GetIPv6(); ipv6.IsGlobalUnicast() {
		prefix = netip.PrefixFrom(ipv6, 64)
	}
	qfi := pdu.GetQosId()
	if qfi < 0 || qfi > 63 {
		return userspace.Config{}, fmt.Errorf("QFI %d is not between 0 and 63", qfi)
	}
	return userspace.Config{UPF: upf, UplinkTEID: pdu.GetTeidUplink(), DownlinkTEID: pdu.GetTeidDownlink(), QFI: uint8(qfi), UE: ue, Prefix: prefix}, nil
}

func startUserspace(name string, pdu *gnbContext.GnbPDUSession, ue *context.UEPDUSession, ip netip.Addr) (datapath, error) {
	config, err := userspaceConfig(pdu, ue)
	if err != nil {
		return nil, err
	}
	session, err := userspace.Open(name, ip, config)
	if err != nil {
		return nil, err
	}
	return userspaceTunnel{session, ue}, nil
}

func (t userspaceTunnel) refresh(pdu *gnbContext.GnbPDUSession, _ string, _ netip.Addr) error {
	config, err := userspaceConfig(pdu, t.session)
	if err == nil {
		t.Update(config)
	}
	return err
}

func (t userspaceTunnel) close() { t.Close() }

func (t userspaceTunnel) solicit(linkLocal netip.Addr) (netip.Addr, error) {
	return t.Solicit(linkLocal)
}
