//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"errors"
	"os"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

type stageTestHook struct {
	attempts int
	fail     bool
}

func (h *stageTestHook) Close() error {
	h.attempts++
	if h.fail {
		return errors.New("injected detach failure")
	}
	return nil
}

func TestChecksumStageCloseRetainsReferencesUntilDetach(t *testing.T) {
	endpoint, relay := &stageTestHook{fail: true}, &stageTestHook{}
	deleted := 0
	s := &checksumStage{tx: &netlink.Veth{}, endpointHook: endpoint, relayHook: relay, deleteLink: func(netlink.Link) error { deleted++; return nil }}
	require.Error(t, s.Close())
	require.Zero(t, relay.attempts)
	require.Zero(t, deleted)
	endpoint.fail, relay.fail = false, true
	require.Error(t, s.Close())
	require.Equal(t, 2, endpoint.attempts)
	require.Nil(t, s.endpointHook)
	require.Zero(t, deleted)
	relay.fail = false
	s.deleteLink = func(netlink.Link) error { deleted++; return errors.New("injected delete failure") }
	require.Error(t, s.Close())
	require.NotNil(t, s.tx)
	require.Nil(t, s.relayHook)
	s.deleteLink = func(netlink.Link) error { deleted++; return nil }
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
	require.Equal(t, 2, endpoint.attempts)
	require.Equal(t, 2, relay.attempts)
	require.Equal(t, 2, deleted)
}

func TestNativeChecksumStageOwnershipAndEffectiveFeatures(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	port, tun, err := userspace.NewTUN("stage-proof")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, port.Close()) })
	program := func(name string, action int32) *ebpf.Program {
		p, e := ebpf.NewProgram(&ebpf.ProgramSpec{Name: name, Type: ebpf.SchedCLS, License: "GPL", Instructions: asm.Instructions{asm.Mov.Imm(asm.R0, action), asm.Return()}})
		require.NoError(t, e)
		t.Cleanup(func() { require.NoError(t, p.Close()) })
		return p
	}
	encap, relay := program("stage_accept", 0), program("stage_drop", 2)
	s, err := attachChecksumStage(tun.Attrs().Index, encap, relay)
	require.NoError(t, err)
	require.NotNil(t, s)
	t.Cleanup(func() { require.NoError(t, s.Close()) })
	tx, peer := s.TransmitIndex(), s.PeerIndex()
	for _, index := range []int{tx, peer} {
		device, e := netlink.LinkByIndex(index)
		require.NoError(t, e)
		require.Equal(t, 65535, device.Attrs().MTU)
		require.NoError(t, disableChecksumStageOffloads(device.Attrs().Name), "effective feature verification after both links are up")
		addresses, e := netlink.AddrList(device, netlink.FAMILY_V4)
		require.NoError(t, e)
		require.Empty(t, addresses)
		t.Logf("owned checksum stage %s index=%d MTU=%d verified checksum/segmentation disabled", device.Attrs().Name, index, device.Attrs().MTU)
	}
	// Simulate LinkAdd's silently failed index resolution: owned name+MAC must
	// still retire the real pair instead of mistaking index zero for absence.
	s.tx.Attrs().Index = 0
	require.NoError(t, s.Close())
	for _, index := range []int{tx, peer} {
		_, e := netlink.LinkByIndex(index)
		var absent netlink.LinkNotFoundError
		require.ErrorAs(t, e, &absent)
	}
	_, err = netlink.LinkByIndex(tun.Attrs().Index)
	require.NoError(t, err, "original caller-owned TUN survives stage cleanup")
	before, err := netlink.LinkList()
	require.NoError(t, err)
	partial, err := attachChecksumStage(2147483647, encap, relay)
	require.Error(t, err, "invalid original endpoint fails after staging")
	require.Nil(t, partial, "successful rollback relinquishes ownership")
	after, err := netlink.LinkList()
	require.NoError(t, err)
	require.Len(t, after, len(before), "failed setup leaves no owned veth behind")
}
