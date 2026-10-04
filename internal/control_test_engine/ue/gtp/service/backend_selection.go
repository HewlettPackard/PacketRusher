// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"my5G-RANTester/config"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"net/netip"
)

var setupEBPFBackend = setupEBPFTunnel
var setupUserspaceBackend = setupUserspaceTunnel

// Setup selection never moves a live endpoint between datapaths. Auto chooses
// once, after constructor rollback, and handover keeps that committed choice.
func setupPortableBackend(ue *context.UEContext, pdu *context.UEPDUSession, peer *gnb.GnbPDUSession, local netip.Addr, requested config.TunnelBackend) error {
	selected, fallback := pdu.TunnelSelection()
	if requested == config.TunnelBackendAuto && pdu.GetTunInterface() != nil && selected == "" {
		return errors.New("existing tunnel backend is unknown; refusing automatic replacement")
	}
	if pdu.GetTunInterface() != nil && selected != "" {
		if requested != config.TunnelBackendAuto && requested != selected {
			return errors.New("release the existing tunnel before changing its backend")
		}
		requested = selected
	}
	if requested == config.TunnelBackendUserspace {
		if err := setupUserspaceBackend(ue, pdu, peer, local); err != nil {
			return err
		}
		pdu.SetTunnelSelection(config.TunnelBackendUserspace, fallback)
		return nil
	}
	if err := setupEBPFBackend(ue, pdu, peer, local); err != nil {
		if requested != config.TunnelBackendAuto || pdu.GetTunInterface() != nil || !errors.Is(err, ebpfgtp.ErrUnavailable) || errors.Is(err, ebpfgtp.ErrCleanupIncomplete) || errors.Is(err, errTunnelRollback) {
			return err
		}
		if nextErr := setupUserspaceBackend(ue, pdu, peer, local); nextErr != nil {
			return fmt.Errorf("eBPF setup failed: %w; userspace fallback failed: %w", err, nextErr)
		}
		pdu.SetTunnelSelection(config.TunnelBackendUserspace, err.Error())
		return nil
	}
	pdu.SetTunnelSelection(config.TunnelBackendEBPF, "")
	return nil
}
