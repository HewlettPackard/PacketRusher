//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// checksumStage gives redirected CHECKSUM_PARTIAL packets a normal device
// transmit boundary. TCX egress runs before validate_xmit_skb; an exclusively
// owned veth with TX offloads disabled completes the retained inner checksum
// before its peer's relay forwards the encapsulated packet to N3.
// The registry owns authorization maps and retires them before Close.
type checksumStage struct {
	mu                      sync.Mutex
	tx, peer                netlink.Link
	txIndex, peerIndex      int
	endpointHook, relayHook closer
	deleteLink              func(netlink.Link) error
}

func (s *checksumStage) TransmitIndex() int { return s.txIndex }
func (s *checksumStage) PeerIndex() int     { return s.peerIndex }

// Close is ordered and retryable. A failed detach retains the devices it could
// still reference. The original TUN and every underlay device belong to callers.
func (s *checksumStage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, hook := range []*closer{&s.endpointHook, &s.relayHook} {
		if *hook != nil {
			if err := (*hook).Close(); err != nil {
				return err
			}
			*hook = nil
		}
	}
	if s.tx != nil {
		if err := s.deleteLink(s.tx); err != nil {
			return err
		}
		s.tx, s.peer = nil, nil
	}
	return nil
}

func attachChecksumStage(endpoint int, encap, relay *ebpf.Program) (result *checksumStage, err error) {
	if endpoint <= 0 || encap == nil || relay == nil {
		return nil, errors.New("checksum stage requires endpoint and both TCX programs")
	}
	var token [5]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	suffix := hex.EncodeToString(token[:])
	tx := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: "prcs" + suffix, MTU: 65535, HardwareAddr: append([]byte{2}, token[:]...)}, PeerName: "prcp" + suffix}
	if err := netlink.LinkAdd(tx); err != nil {
		return nil, fmt.Errorf("create owned checksum veth: %w", err)
	}
	s := &checksumStage{tx: tx, txIndex: tx.Attrs().Index, deleteLink: deleteChecksumLink}
	defer func() {
		if err != nil {
			if cleanup := s.Close(); cleanup != nil {
				err = errors.Join(err, fmt.Errorf("retire partial checksum stage: %w", cleanup))
				// Return ownership for a later retry instead of losing live objects.
				result = s
			} else {
				result = nil
			}
		}
	}()
	// Return s even on a partial failure; callers must retain it if cleanup fails.
	finish := func(e error) (*checksumStage, error) { return s, e }
	// LinkAdd's internal index lookup can fail silently after a successful
	// creation. Resolve it explicitly while retaining name/MAC cleanup ownership.
	resolved, err := netlink.LinkByName(tx.Attrs().Name)
	if err != nil {
		return finish(err)
	}
	if resolved.Type() != "veth" || !bytes.Equal(resolved.Attrs().HardwareAddr, tx.Attrs().HardwareAddr) {
		return finish(errors.New("created checksum veth identity changed during setup"))
	}
	s.tx = resolved
	s.txIndex = s.tx.Attrs().Index
	if s.txIndex <= 0 {
		return finish(errors.New("created checksum veth has no interface index"))
	}
	s.peer, err = netlink.LinkByName(tx.PeerName)
	if err != nil {
		return finish(err)
	}
	s.peerIndex = s.peer.Attrs().Index
	for _, device := range []netlink.Link{s.tx, s.peer} {
		if err = netlink.LinkSetMTU(device, 65535); err != nil {
			return finish(err)
		}
		if err = disableChecksumStageOffloads(device.Attrs().Name); err != nil {
			return finish(err)
		}
	}
	s.relayHook, err = link.AttachTCX(link.TCXOptions{Interface: s.PeerIndex(), Program: relay, Attach: ebpf.AttachTCXIngress, Anchor: link.Tail()})
	if err != nil {
		return finish(fmt.Errorf("attach checksum relay: %w", err))
	}
	for _, device := range []netlink.Link{s.peer, s.tx} {
		if err = netlink.LinkSetUp(device); err != nil {
			return finish(err)
		}
	}
	s.endpointHook, err = link.AttachTCX(link.TCXOptions{Interface: endpoint, Program: encap, Attach: ebpf.AttachTCXEgress, Anchor: link.Tail()})
	if err != nil {
		return finish(fmt.Errorf("attach checksum endpoint: %w", err))
	}
	return finish(nil)
}

func deleteChecksumLink(owned netlink.Link) error {
	var actual netlink.Link
	var err error
	if owned.Attrs().Index > 0 {
		actual, err = netlink.LinkByIndex(owned.Attrs().Index)
	} else {
		actual, err = netlink.LinkByName(owned.Attrs().Name)
	}
	var absent netlink.LinkNotFoundError
	if errors.As(err, &absent) || errors.Is(err, unix.ENODEV) {
		return nil
	}
	if err != nil {
		return err
	}
	if actual.Type() != "veth" || actual.Attrs().Name != owned.Attrs().Name || !bytes.Equal(actual.Attrs().HardwareAddr, owned.Attrs().HardwareAddr) {
		return errors.New("checksum veth identity changed; refusing to delete another device")
	}
	return netlink.LinkDel(actual)
}

// Linux ifreq's union contains struct ifmap (two native longs and eight bytes).
// Keeping the data pointer typed preserves its lifetime through the ioctl.
type checksumIfreq struct {
	name [unix.IFNAMSIZ]byte
	data unsafe.Pointer
	_    [8 + unix.SizeofPtr]byte
}

func checksumEthtool(fd int, name string, data unsafe.Pointer) error {
	if len(name) == 0 || len(name) >= unix.IFNAMSIZ {
		return errors.New("invalid checksum stage interface name")
	}
	ifr := checksumIfreq{data: data}
	copy(ifr.name[:], name)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCETHTOOL, uintptr(unsafe.Pointer(&ifr)))
	runtime.KeepAlive(ifr)
	if errno != 0 {
		return errno
	}
	return nil
}

type checksumFeatureBlock struct{ Available, Requested, Active, Fixed uint32 }
type checksumFeatures struct {
	Command, Size uint32
	Blocks        [16]checksumFeatureBlock
}
type checksumSetFeatures struct {
	Command, Size uint32
	Blocks        [16]struct{ Valid, Requested uint32 }
}

func disableChecksumStageOffloads(name string) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	features := checksumFeatures{Command: unix.ETHTOOL_GFEATURES, Size: 16}
	if err := checksumEthtool(fd, name, unsafe.Pointer(&features)); err != nil {
		return err
	}
	if features.Size == 0 || features.Size > uint32(len(features.Blocks)) {
		return errors.New("unsupported checksum stage feature bitmap")
	}
	request := checksumSetFeatures{Command: unix.ETHTOOL_SFEATURES, Size: features.Size}
	for i := uint32(0); i < features.Size; i++ {
		request.Blocks[i].Valid = features.Blocks[i].Available
	}
	// This device exists only to materialize checksums and segmentation. Disable
	// every changeable feature, including future UDP/SCTP segmentation flags.
	if err := checksumEthtool(fd, name, unsafe.Pointer(&request)); err != nil {
		return err
	}
	if err := checksumEthtool(fd, name, unsafe.Pointer(&features)); err != nil {
		return err
	}
	for _, b := range features.Blocks[:features.Size] {
		if b.Available&b.Active != 0 {
			return fmt.Errorf("checksum stage %s retains changeable offloads", name)
		}
	}
	// Check the kernel's effective checksum/TSO/GSO masks as well as readback of
	// requested features: successful feature requests alone are insufficient.
	for _, command := range []uint32{unix.ETHTOOL_GTXCSUM, unix.ETHTOOL_GTSO, unix.ETHTOOL_GGSO} {
		value := struct{ Command, Data uint32 }{Command: command}
		if err := checksumEthtool(fd, name, unsafe.Pointer(&value)); err != nil {
			return err
		}
		if value.Data != 0 {
			return fmt.Errorf("checksum stage %s retains effective offload %#x", name, command)
		}
	}
	return nil
}
