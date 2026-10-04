/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package ebpfgtp short-cuts the userspace GTP-U backend in the kernel: eBPF
// programs encapsulate what a UE's TUN device transmits and decapsulate what its
// N3 interface receives. What they leave alone still goes through userspace.
package ebpfgtp

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Built from gtpu.c by "make ebpf".
//
//go:embed gtpu_bpfel.o
var object []byte

// uplink and downlinkKey mirror the structures of gtpu.c.
type uplink struct {
	Local, Peer, TEID [4]byte
	N3                uint32
	UE                [4]byte
	Prefix            [8]byte
	QFI               uint8
	_                 [3]byte
}

type downlinkKey struct{ Local, TEID [4]byte }

type tunnel struct {
	index, n3 int
	uplink    link.Link
	key       downlinkKey
}

var (
	mu      sync.Mutex
	objects *ebpf.Collection
	stage   netlink.Link // its near end
	encap   link.Link
	decaps  = map[int]link.Link{}  // by interface towards a UPF
	tunnels = map[string]*tunnel{} // by TUN device name
)

// The stage is a veth pair, prs<n>a and prs<n>b.
const stagePrefix = "prs"

// Load prepares the fast path: it loads the programs and creates the stage, a veth
// pair without checksum offload. The UEs' packets cross it before their
// encapsulation: transmitting through it splits the large segments of their TCP stack
// and completes their checksums.
// It fails without TCX (Linux 6.6) or without the rights to load eBPF programs.
func Load() error {
	mu.Lock()
	defer mu.Unlock()
	if objects != nil {
		return nil
	}
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(object))
	if err != nil {
		return err
	}
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("loading the eBPF programs: %w", err)
	}
	if err := createStage(collection); err != nil {
		collection.Close()
		return err
	}
	objects = collection
	return nil
}

// createStage creates the stage under the first free name, a process ID being the
// same in every container, and attaches to it at once: a stage with nothing attached
// is one that its process left behind.
func createStage(collection *ebpf.Collection) error {
	removeStaleStages()
	var pair *netlink.Veth
	err := error(unix.EEXIST)
	for n := 0; errors.Is(err, unix.EEXIST); n++ {
		name := stagePrefix + strconv.Itoa(n)
		// No UE packet is larger than its MTU, whatever that of N3.
		pair = &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: name + "a", MTU: 65535}, PeerName: name + "b"}
		err = netlink.LinkAdd(pair)
	}
	if err != nil {
		return fmt.Errorf("creating the stage: %w", err)
	}
	far, err := netlink.LinkByName(pair.PeerName)
	if err == nil {
		encap, err = link.AttachTCX(link.TCXOptions{Interface: far.Attrs().Index, Program: collection.Programs["encap"], Attach: ebpf.AttachTCXIngress})
	}
	if err != nil {
		netlink.LinkDel(pair)
		return fmt.Errorf("attaching to the stage: %w", err)
	}
	near, err := netlink.LinkByName(pair.Name)
	if err == nil {
		err = errors.Join(ethtool(pair.Name, unix.ETHTOOL_STXCSUM, 0), netlink.LinkSetUp(near), netlink.LinkSetUp(far),
			collection.Variables["stage"].Set(uint32(near.Attrs().Index)))
	}
	if err != nil {
		encap.Close()
		netlink.LinkDel(pair)
		return err
	}
	steer(pair.PeerName)
	stage = near
	return nil
}

// removeStaleStages deletes the stages of processes that ended without Close: what
// was attached to them ended with their process.
func removeStaleStages() {
	links, _ := netlink.LinkList()
	for _, l := range links {
		number, isFarEnd := strings.CutSuffix(strings.TrimPrefix(l.Attrs().Name, stagePrefix), "b")
		if _, err := strconv.Atoi(number); err != nil || !isFarEnd || l.Type() != "veth" {
			continue
		}
		attached, err := link.QueryPrograms(link.QueryOptions{Target: l.Attrs().Index, Attach: ebpf.AttachTCXIngress})
		if err == nil && len(attached.Programs) == 0 {
			netlink.LinkDel(l)
		}
	}
}

// Attach short-cuts the tunnel of the TUN device name between the N3 address
// local and the UPF. Attaching it again applies the tunnel's new identifiers.
func Attach(name string, local netip.Addr, c userspace.Config) error {
	mu.Lock()
	defer mu.Unlock()
	// The interface that leads to the UPF, the loopback one when it is on this host:
	// not always the one that holds the N3 address.
	routes, err := netlink.RouteGetWithOptions(c.UPF.AsSlice(), &netlink.RouteGetOptions{SrcAddr: local.AsSlice()})
	if err != nil || len(routes) == 0 {
		return fmt.Errorf("no route to the UPF %s: %v", c.UPF, err)
	}
	n3 := routes[0].LinkIndex
	if decaps[n3] == nil {
		decaps[n3], err = link.AttachTCX(link.TCXOptions{Interface: n3, Program: objects.Programs["decap"], Attach: ebpf.AttachTCXIngress})
		if err != nil {
			delete(decaps, n3)
			return fmt.Errorf("attaching to the N3 interface: %w", err)
		}
	}
	t := tunnels[name]
	if t == nil {
		tun, err := netlink.LinkByName(name)
		if err != nil {
			release(n3)
			return err
		}
		steer(name)
		t = &tunnel{index: tun.Attrs().Index, n3: n3}
		tunnels[name] = t
	} else {
		objects.Maps["downlinks"].Delete(t.key)
	}
	previous := t.n3
	t.n3 = n3
	release(previous)
	t.key = downlinkKey{Local: local.As4(), TEID: bigEndian(c.DownlinkTEID)}
	up := uplink{Local: local.As4(), Peer: c.UPF.As4(), TEID: bigEndian(c.UplinkTEID), N3: uint32(t.n3), QFI: c.QFI}
	if c.UE.Is4() {
		up.UE = c.UE.As4()
	}
	if c.Prefix.IsValid() {
		prefix := c.Prefix.Addr().As16()
		copy(up.Prefix[:], prefix[:8])
	}
	if err := errors.Join(objects.Maps["uplinks"].Put(uint32(t.index), up), objects.Maps["downlinks"].Put(t.key, uint32(t.index))); err != nil {
		return err
	}
	if t.uplink == nil {
		var err error
		t.uplink, err = link.AttachTCX(link.TCXOptions{Interface: t.index, Program: objects.Programs["uplink"], Attach: ebpf.AttachTCXEgress})
		if err != nil {
			return fmt.Errorf("attaching to %s: %w", name, err)
		}
	}
	return nil
}

// release detaches from the interface n3 once no tunnel is left on it.
func release(n3 int) {
	for _, t := range tunnels {
		if t.n3 == n3 {
			return
		}
	}
	if decap := decaps[n3]; decap != nil {
		decap.Close()
		delete(decaps, n3)
	}
}

// Detach leaves the tunnel of the TUN device name to userspace again.
func Detach(name string) {
	mu.Lock()
	defer mu.Unlock()
	t := tunnels[name]
	if t == nil {
		return
	}
	delete(tunnels, name)
	objects.Maps["uplinks"].Delete(uint32(t.index))
	objects.Maps["downlinks"].Delete(t.key)
	if t.uplink != nil {
		t.uplink.Close()
	}
	release(t.n3)
}

// Close removes the stage and the programs. The tunnels must be detached.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if objects == nil {
		return
	}
	encap.Close()
	objects.Close()
	objects = nil
	netlink.LinkDel(stage)
}

// steer spreads what one of our devices receives over the CPUs this process may use:
// otherwise the CPU of whoever transmits to it, an application or the N3 interface,
// also does the encapsulation or the reception of the UE's packets.
func steer(device string) {
	var cpus unix.CPUSet
	if unix.SchedGetaffinity(0, &cpus) != nil {
		return
	}
	mask := new(big.Int)
	for cpu := range len(cpus) * 64 {
		if cpus.IsSet(cpu) {
			mask.SetBit(mask, cpu, 1)
		}
	}
	bitmap := mask.Text(16) // sysfs takes it in groups of 32 bits
	for i := len(bitmap) - 8; i > 0; i -= 8 {
		bitmap = bitmap[:i] + "," + bitmap[i:]
	}
	_ = os.WriteFile("/sys/class/net/"+device+"/queues/rx-0/rps_cpus", []byte(bitmap), 0)
}

// ethtool runs one of the ethtool commands that set a single value on a device.
func ethtool(device string, command, value uint32) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	data := [2]uint32{command, value}
	request := struct {
		name [unix.IFNAMSIZ]byte
		data unsafe.Pointer
		_    [16]byte
	}{data: unsafe.Pointer(&data)}
	copy(request.name[:], device)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCETHTOOL, uintptr(unsafe.Pointer(&request))); errno != 0 {
		return fmt.Errorf("ethtool %#x on %s: %w", command, device, errno)
	}
	return nil
}

func bigEndian(value uint32) (b [4]byte) {
	binary.BigEndian.PutUint32(b[:], value)
	return
}
