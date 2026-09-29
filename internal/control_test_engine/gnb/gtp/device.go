/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package gtp

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/free5gc/go-gtp5gnl"
	gtpLink "github.com/free5gc/go-gtp5gnl/linkcmd"
	gtpTunnel "github.com/free5gc/go-gtp5gnl/tuncmd"
	"github.com/khirono/go-nl"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// A gtp5g device owns the GTP-U socket on the address it is created with, and an
// address:2152 pair can only be bound once. Giving every UE its own device therefore
// forced one gNB per UE -- which is why --tunnel required --dedicatedGnb.
//
// A real gNB does not work that way: one N3 endpoint carries thousands of UEs and
// tells them apart by TEID. gtp5g supports exactly that, since PDRs and FARs are
// per-device rules and a device may hold many. Device is that arrangement: one
// device per gNB, created and closed by the gNB, with each UE adding and removing
// only its own PDRs, FARs and address.
type Device struct {
	name      string
	stop      chan bool
	closeOnce sync.Once

	// mu guards nextUE and free. IDs must be unique within a device, so they are
	// handed out centrally rather than hardcoded per UE as the per-UE-device design
	// could afford to do, and returned when a UE's rules are removed so that a long
	// --loop run does not exhaust them.
	mu     sync.Mutex
	nextUE uint32
	free   []uint32

	// qerMu guards qers, the device's QERs by QFI. UEs on one gNB may be given
	// different QFIs, and a PDR must reference a QER carrying its own flow's QFI, so
	// each QFI gets one QER, created by the first UE that needs it. gtp5g refuses a
	// QER that already exists, so creating one per UE would fail for every UE after
	// the first. An identifier is reserved for a QFI at its first attempt and never
	// handed to another, so a create that the kernel applied but did not acknowledge
	// cannot collide with a later QFI's.
	qerMu   sync.Mutex
	qers    map[int64]*qerState
	nextQER uint32

	// One netlink client per device, held for the run, rather than the fresh
	// conn/mux/client the tuncmd.Cmd* wrappers build for every single call. It installs
	// every UE's rules and, when PR_VERIFY_RULES=1, reads PDRs back: a rule can fail to
	// install without the kernel or the library saying so, and a UE with no rules still
	// looks configured -- it has an address, a policy rule and a route, and its packets
	// leave userspace without error. They are simply dropped.
	clientMu sync.Mutex
	link     *gtp5gnl.Link
	conn     *nl.Conn
	mux      *nl.Mux
	client   *gtp5gnl.Client
}

// NewDevice creates the GTP-U device a gNB shares among its UEs, on its N3 address.
// A device of the same name left behind by an earlier run still owns the GTP-U socket
// on that address, so it is removed first. Only this gNB's device is touched: several
// PacketRusher processes can share a host, one per gNB, and removing every device
// named like ours would tear down the tunnels of the siblings that are running.
func NewDevice(gnbIP netip.Addr) (*Device, error) {
	name := deviceName(gnbIP)

	if _, err := netlink.LinkByName(name); err == nil {
		if err := gtpLink.CmdDel(name); err != nil {
			return nil, fmt.Errorf("unable to remove stale GTP-U device %s: %w", name, err)
		}
		log.Info("[GNB][GTP] Removed stale shared GTP-U device ", name, " from a previous run")
	}

	d := &Device{name: name, stop: make(chan bool)}

	go func() {
		// Does not return while the GTP-U socket is open, so it owns this goroutine
		// for the lifetime of the gNB rather than of one UE.
		if err := gtpLink.CmdAddWithStopCh(name, 1, 131072, gnbIP.String(), "", d.stop); err != nil {
			log.Error("[GNB][GTP] shared GTP device ", name, " ended: ", err)
		}
	}()

	if err := waitForLink(name, 5*time.Second); err != nil {
		close(d.stop)
		return nil, err
	}

	d.openClient()

	log.Info(fmt.Sprintf("[GNB][GTP] shared GTP-U device %s created on %s; every UE of this gNB will use it", name, gnbIP))

	return d, nil
}

// Name is the device's interface name.
func (d *Device) Name() string {
	return d.name
}

// Close releases the GTP-U socket, removes the device and closes its netlink client.
// Only the gNB calls it, when it terminates.
func (d *Device) Close() {
	d.closeOnce.Do(func() {
		d.clientMu.Lock()
		if d.conn != nil {
			d.conn.Close()
			d.mux.Close()
		}
		d.client, d.link, d.conn, d.mux = nil, nil, nil, nil
		d.clientMu.Unlock()

		close(d.stop)
		_ = gtpLink.CmdDel(d.name)
	})
}

// openClient prepares the device's long-lived netlink client.
func (d *Device) openClient() {
	mux, err := nl.NewMux()
	if err != nil {
		log.Warn("[GNB][GTP] netlink client unavailable (mux): ", err)
		return
	}

	go mux.Serve()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		mux.Close()
		log.Warn("[GNB][GTP] netlink client unavailable (conn): ", err)

		return
	}

	client, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		conn.Close()
		mux.Close()
		log.Warn("[GNB][GTP] netlink client unavailable (client): ", err)

		return
	}

	link, err := gtp5gnl.GetLink(d.name)
	if err != nil {
		conn.Close()
		mux.Close()
		log.Warn("[GNB][GTP] netlink client unavailable (link): ", err)

		return
	}

	d.mux = mux
	d.conn = conn
	d.client = client
	d.link = link
}

// PDRInstalled reports whether the rule really reached the datapath. It answers true
// when there is no client, so a device without one degrades to the old behaviour
// rather than declaring every UE broken.
func (d *Device) PDRInstalled(id uint32) bool {
	d.clientMu.Lock()
	defer d.clientMu.Unlock()

	if d.client == nil || d.link == nil {
		return true
	}

	if _, err := gtp5gnl.GetPDR(d.client, d.link, int(id)); err != nil {
		return false
	}

	return true
}

// Each tuncmd.Cmd* call opens a netlink socket, starts a mux goroutine, builds a
// client and looks the link up by name, then tears all of it down again -- about five
// such lifecycles per UE across its PDRs, FARs and QER. That is most of the cost of
// bringing a tunnel up, and the socket churn is why doing this in parallel produced
// "bad file descriptor" and forced the concurrency bound down to four.
//
// These parse the same argument vectors the Cmd* wrappers accept, then issue the create
// on the device's long-lived client. They fall back to the wrapper when no client is
// available, so a device whose client could not be opened still works exactly as before.

// addRule issues one create on the shared client, or falls back to the wrapper.
func (d *Device) addRule(
	args []string,
	parse func([]string) ([]nl.Attr, error),
	create func(*gtp5gnl.Client, *gtp5gnl.Link, gtp5gnl.OID, []nl.Attr) error,
	fallback func([]string) error,
) error {
	d.clientMu.Lock()
	defer d.clientMu.Unlock()

	if d.client == nil || d.link == nil {
		return fallback(args)
	}

	if len(args) < 2 {
		return fmt.Errorf("too few parameters for a rule on %s", d.name)
	}

	oid, err := gtpTunnel.ParseOID(args[1])
	if err != nil {
		return err
	}

	attrs, err := parse(args[2:])
	if err != nil {
		return err
	}

	return create(d.client, d.link, oid, attrs)
}

// AddPDR installs a PDR through the device's own client.
func (d *Device) AddPDR(args []string) error {
	return d.addRule(args, gtpTunnel.ParsePDROptions, gtp5gnl.CreatePDROID, gtpTunnel.CmdAddPDR)
}

// AddFAR installs a FAR through the device's own client.
func (d *Device) AddFAR(args []string) error {
	return d.addRule(args, gtpTunnel.ParseFAROptions, gtp5gnl.CreateFAROID, gtpTunnel.CmdAddFAR)
}

// AddQER installs a QER through the device's own client.
func (d *Device) AddQER(args []string) error {
	return d.addRule(args, gtpTunnel.ParseQEROptions, gtp5gnl.CreateQEROID, gtpTunnel.CmdAddQER)
}

// qerState is one QFI's QER on a device: its reserved identifier, and how the
// attempts to create it have gone.
type qerState struct {
	id       uint32
	created  bool
	failures int
}

// maxQERAttempts bounds how many UEs try to create one QFI's QER. Past it, UEs with
// that QFI get tunnels whose uplink carries no QFI marking, and that is logged once.
const maxQERAttempts = 3

// EnsureQER returns the identifier of this device's QER for qfi, calling create the
// first time that QFI is seen. It reports false when the QER is not usable, so the
// caller leaves it unreferenced; the next UE with that QFI tries again with the same
// identifier, up to maxQERAttempts in all. A create refused because the QER already
// exists counts as done: the identifier is reserved for this QFI, so the QER there
// can only be this QFI's, from an attempt whose acknowledgement was lost.
func (d *Device) EnsureQER(qfi int64, create func(id uint32) error) (uint32, bool) {
	d.qerMu.Lock()
	defer d.qerMu.Unlock()

	if d.qers == nil {
		d.qers = make(map[int64]*qerState)
	}

	q, ok := d.qers[qfi]
	if !ok {
		// QFIs are six bits, so a device never holds more than 64 of these.
		d.nextQER++
		q = &qerState{id: d.nextQER}
		d.qers[qfi] = q
	}

	if q.created {
		return q.id, true
	}

	if q.failures >= maxQERAttempts {
		return 0, false
	}

	if err := create(q.id); err != nil && !errors.Is(err, syscall.EEXIST) {
		q.failures++
		if q.failures < maxQERAttempts {
			log.Error("[UE][GTP] Unable to create the shared QER for QFI ", qfi, ", will retry: ", err)
		} else {
			log.Error("[UE][GTP] Unable to create the shared QER for QFI ", qfi, " after ", q.failures,
				" attempts; UEs with this QFI will carry no QFI marking on the uplink: ", err)
		}

		return 0, false
	}

	q.created = true

	return q.id, true
}

// RuleIDs are the rule identifiers one UE occupies on a device. PDR and FAR are
// separate identifier spaces in gtp5g, so the same numbers may repeat across them.
// QERs are per QFI, not per UE, and so are not among them; see EnsureQER.
type RuleIDs struct {
	Slot    uint32
	PDRDown uint32
	PDRUp   uint32
	FARDown uint32
	FARUp   uint32
}

// DedicatedQERID is the QER a UE uses on a device it owns outright.
const DedicatedQERID = 1

// deviceNamePrefix names every device this mode creates.
const deviceNamePrefix = "valgnb"

// deviceName derives a device name from the gNB's N3 address. Interface names are
// capped at 15 characters, and the hex form of an IPv4 address keeps this at 14 while
// staying unique per gNB.
func deviceName(gnbIP netip.Addr) string {
	unmapped := gnbIP.Unmap()
	if !unmapped.Is4() {
		return deviceNamePrefix + "shared"
	}

	v4 := unmapped.As4()

	return fmt.Sprintf("%s%02x%02x%02x%02x", deviceNamePrefix, v4[0], v4[1], v4[2], v4[3])
}

// waitForLink polls until the device the creating goroutine asked for exists. The
// per-UE-device code slept a fixed second for this; at scale that is both slower than
// necessary and not actually a guarantee.
func waitForLink(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := netlink.LinkByName(name); err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("shared GTP device %s did not appear within %s", name, timeout)
		}

		time.Sleep(20 * time.Millisecond)
	}
}

// maxSlots is how many UEs one device can hold rules for at once. gtp5g's PDR
// identifier is 16 bits and each UE takes two, 2n+1 and 2n+2, so the last usable
// slot is the one whose uplink PDR is 65534.
const maxSlots = 32767

// ErrDeviceFull reports that a device has no rule identifiers left for another UE.
var ErrDeviceFull = errors.New("shared GTP-U device has no rule identifiers left")

// Take reserves one UE's worth of rule identifiers, reusing a slot a departed UE gave
// back before opening a new one.
func (d *Device) Take() (RuleIDs, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var n uint32
	if last := len(d.free) - 1; last >= 0 {
		n = d.free[last]
		d.free = d.free[:last]
	} else {
		if d.nextUE >= maxSlots {
			return RuleIDs{}, ErrDeviceFull
		}
		n = d.nextUE
		d.nextUE++
	}

	return idsForSlot(n), nil
}

// inUse reports how many slots are taken and not yet given back.
func (d *Device) inUse() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return int(d.nextUE) - len(d.free)
}

// WaitIdle waits for every UE to give its rules back, and reports how many still hold
// them. It gives up once stall passes with no further slot given back, or once limit
// passes in all: a slot whose rules could not be removed is never given back, and in
// --loop mode UEs register again after they terminate, so zero may never come.
func (d *Device) WaitIdle(stall, limit time.Duration) int {
	start := time.Now()
	lowest, lowestAt := d.inUse(), start
	for {
		n := d.inUse()
		now := time.Now()
		if n < lowest {
			lowest, lowestAt = n, now
		}

		if n == 0 || now.Sub(lowestAt) >= stall || now.Sub(start) >= limit {
			return n
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// giveBack makes a slot available to take again. Only call it once the slot's rules
// are gone from the device, or the next UE's creates collide with them.
func (d *Device) giveBack(ids RuleIDs) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.free = append(d.free, ids.Slot)
}

// Release removes one UE's PDRs and FARs from the device and returns their slot. The
// QERs are the device's, not the UE's, and stay. If a rule could not be removed the slot
// is not reused: its identifiers are still occupied.
func (d *Device) Release(ids RuleIDs) {
	if err := d.removeRules(ids); err != nil {
		log.Warn("[UE][GTP] Unable to remove the rules of slot ", ids.Slot, " on ", d.name, "; not reusing it: ", err)
		return
	}

	d.giveBack(ids)
}

// removeRules deletes a UE's PDRs, then the FARs they reference. A rule that is already
// absent -- say, because setup failed before creating it -- counts as removed.
func (d *Device) removeRules(ids RuleIDs) error {
	d.clientMu.Lock()
	defer d.clientMu.Unlock()

	type rule struct {
		id     uint32
		remove func(*gtp5gnl.Client, *gtp5gnl.Link, int) error
		cmd    func([]string) error
	}

	var errs []error
	for _, r := range []rule{
		{ids.PDRDown, gtp5gnl.RemovePDR, gtpTunnel.CmdDeletePDR},
		{ids.PDRUp, gtp5gnl.RemovePDR, gtpTunnel.CmdDeletePDR},
		{ids.FARDown, gtp5gnl.RemoveFAR, gtpTunnel.CmdDeleteFAR},
		{ids.FARUp, gtp5gnl.RemoveFAR, gtpTunnel.CmdDeleteFAR},
	} {
		var err error
		if d.client != nil && d.link != nil {
			err = r.remove(d.client, d.link, int(r.id))
		} else {
			err = r.cmd([]string{d.name, strconv.FormatUint(uint64(r.id), 10)})
		}

		if err != nil && !errors.Is(err, syscall.ENOENT) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// RemoveAddress takes one UE's address off the device.
func (d *Device) RemoveAddress(ueIP string) {
	dev, err := netlink.LinkByName(d.name)
	if err != nil {
		return
	}

	ip := net.ParseIP(ueIP).To4()
	if ip == nil {
		return
	}

	_ = netlink.AddrDel(dev, &netlink.Addr{IPNet: &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}})
}

// idsForSlot maps a slot to its rule identifiers.
func idsForSlot(n uint32) RuleIDs {
	return RuleIDs{
		Slot:    n,
		PDRDown: 2*n + 1,
		PDRUp:   2*n + 2,
		FARDown: 2*n + 1,
		FARUp:   2*n + 2,
	}
}

// DedicatedRuleIDs are the identifiers used when a UE owns its device outright, kept
// at their original values so that mode behaves exactly as before.
func DedicatedRuleIDs() RuleIDs {
	return RuleIDs{PDRDown: 1, PDRUp: 2, FARDown: 1, FARUp: 2}
}
