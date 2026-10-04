// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"net/netip"
	"testing"
)

func TestProductionBackendDispatchFallbackAndStickyHandover(t *testing.T) {
	oldEBPF, oldUserspace := setupEBPFBackend, setupUserspaceBackend
	t.Cleanup(func() { setupEBPFBackend, setupUserspaceBackend = oldEBPF, oldUserspace })
	for _, tc := range []struct {
		name      string
		selection config.TunnelBackend
		failure   error
		fallback  bool
	}{
		{"default unavailable", "", errors.Join(ebpfgtp.ErrUnavailable, unix.EPERM), true},
		{"auto unavailable", config.TunnelBackendAuto, errors.Join(ebpfgtp.ErrUnavailable, unix.EOPNOTSUPP), true},
		{"strict", config.TunnelBackendEBPF, errors.Join(ebpfgtp.ErrUnavailable, unix.EPERM), false},
		{"cleanup incomplete", config.TunnelBackendAuto, errors.Join(ebpfgtp.ErrUnavailable, ebpfgtp.ErrCleanupIncomplete, unix.EIO), false},
		{"policy rejection", config.TunnelBackendAuto, unix.EACCES, false},
		{"ownership collision", config.TunnelBackendAuto, unix.EADDRINUSE, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ue, pdu := sharedSetupUE(t, 1)
			ue.TunnelBackend = tc.selection
			source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
			link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: "selected"}}
			ebpfCalls, portableCalls := 0, 0
			setupEBPFBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
				ebpfCalls++
				return tc.failure
			}
			setupUserspaceBackend = func(_ *context.UEContext, p *context.UEPDUSession, _ *gnb.GnbPDUSession, _ netip.Addr) error {
				portableCalls++
				p.SetTunInterface(link)
				p.SetTunnelCleanup(func(bool) {})
				return nil
			}
			SetupGtpInterface(ue, source)
			require.Equal(t, 1, ebpfCalls)
			if !tc.fallback {
				require.Zero(t, portableCalls)
				require.Nil(t, pdu.GetTunInterface())
				return
			}
			require.Equal(t, 1, portableCalls)
			backend, cause := pdu.TunnelSelection()
			require.Equal(t, config.TunnelBackendUserspace, backend)
			require.Contains(t, cause, tc.failure.Error())
			target := ebpfTestMessage(t, "10.88.0.3", 1002, 2002)
			SetupGtpInterface(ue, target)
			require.Equal(t, 1, ebpfCalls, "handover must not retry a different datapath")
			require.Equal(t, 2, portableCalls)
			require.Equal(t, target.GnbIp, pdu.GetGnbIp())
			_, nextCause := pdu.TunnelSelection()
			require.Equal(t, cause, nextCause)
			pdu.ReleaseTunnel()
			backend, cause = pdu.TunnelSelection()
			require.Empty(t, backend)
			require.Empty(t, cause)
		})
	}
}

func TestEBPFReleaseReportsEveryRetainedOwnershipBoundary(t *testing.T) {
	for _, phase := range []string{"endpoint", "route", "rule", "VRF", "claim"} {
		t.Run(phase, func(t *testing.T) {
			_, pdu := sharedSetupUE(t, 1)
			failure := errors.New("injected " + phase + " retirement failure")
			oldRoute, oldRule, oldLink := routeDel, ruleDel, deleteTunnelLink
			routeDel = func(*netlink.Route) error {
				if phase == "route" {
					return failure
				}
				return nil
			}
			ruleDel = func(*netlink.Rule) error {
				if phase == "rule" {
					return failure
				}
				return nil
			}
			deleteTunnelLink = func(netlink.Link) error {
				if phase == "VRF" {
					return failure
				}
				return nil
			}
			t.Cleanup(func() { routeDel, ruleDel, deleteTunnelLink = oldRoute, oldRule, oldLink })
			a := newRoutingTableAllocator(func(*netlink.Route) error { return nil }, func(*netlink.Route) error {
				if phase == "claim" {
					return failure
				}
				return nil
			}, func(uint32) (bool, error) { return false, nil })
			table, _, err := a.reserve(pdu)
			require.NoError(t, err)
			port := &ebpfFailurePort{}
			if phase == "endpoint" {
				port.err = failure
			}
			tunnel := &ebpfTunnel{port: port, table: table, route: &netlink.Route{}, rule: netlink.NewRule(), vrf: &netlink.Vrf{}}
			t.Cleanup(func() { quarantinedEBPFEndpoints.Delete(tunnel) })
			err = tunnel.release()
			require.ErrorIs(t, err, ebpfgtp.ErrCleanupIncomplete)
			require.ErrorIs(t, err, failure)
			require.ErrorIs(t, tunnel.release(), failure, "the second call cannot convert incomplete cleanup to success")
			require.Empty(t, a.free)
		})
	}
}

func TestAutoNeverReplacesLiveBackendOnUnavailableHandover(t *testing.T) {
	oldEBPF, oldUserspace := setupEBPFBackend, setupUserspaceBackend
	t.Cleanup(func() { setupEBPFBackend, setupUserspaceBackend = oldEBPF, oldUserspace })
	ue, pdu := sharedSetupUE(t, 1)
	ue.TunnelBackend = config.TunnelBackendAuto
	source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
	target := ebpfTestMessage(t, "10.88.0.3", 1002, 2002)
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: "committed"}}
	pdu.SetTunInterface(link)
	pdu.SetGnbIp(source.GnbIp)
	pdu.GnbPduSession = source.GNBPduSessions[0]
	pdu.SetTunnelSelection(config.TunnelBackendEBPF, "")
	setupEBPFBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
		return errors.Join(ebpfgtp.ErrUnavailable, unix.EOPNOTSUPP)
	}
	setupUserspaceBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
		t.Fatal("live handover must not migrate to userspace")
		return nil
	}
	SetupGtpInterface(ue, target)
	require.Same(t, link, pdu.GetTunInterface())
	require.Equal(t, source.GnbIp, pdu.GetGnbIp())
	require.Same(t, source.GNBPduSessions[0], pdu.GnbPduSession)
	backend, reason := pdu.TunnelSelection()
	require.Equal(t, config.TunnelBackendEBPF, backend)
	require.Empty(t, reason)
}

func TestFailedFallbackPreservesBothSetupErrorsAndNoSelection(t *testing.T) {
	oldEBPF, oldUserspace := setupEBPFBackend, setupUserspaceBackend
	t.Cleanup(func() { setupEBPFBackend, setupUserspaceBackend = oldEBPF, oldUserspace })
	ue, pdu := sharedSetupUE(t, 1)
	source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
	setupEBPFBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
		return errors.Join(ebpfgtp.ErrUnavailable, unix.EPERM)
	}
	setupUserspaceBackend = func(*context.UEContext, *context.UEPDUSession, *gnb.GnbPDUSession, netip.Addr) error {
		return unix.ENOSPC
	}
	err := setupPortableBackend(ue, pdu, source.GNBPduSessions[0], source.GnbIp, config.TunnelBackendAuto)
	require.ErrorIs(t, err, ebpfgtp.ErrUnavailable)
	require.ErrorIs(t, err, unix.EPERM)
	require.ErrorIs(t, err, unix.ENOSPC)
	backend, reason := pdu.TunnelSelection()
	require.Empty(t, backend)
	require.Empty(t, reason)
	require.Nil(t, pdu.GetTunInterface())
}
