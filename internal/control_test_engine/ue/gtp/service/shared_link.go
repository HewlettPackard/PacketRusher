/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
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
// per-device rules and a device may hold many. sharedLink is that arrangement: one
// device per gNB, with each UE contributing its own PDR/FAR/QER set.
type sharedLink struct {
	name string
	stop chan bool

	// mu guards nextUE and free. IDs must be unique within a device, so they are
	// handed out centrally rather than hardcoded per UE as the per-UE-device design
	// could afford to do, and returned when a UE's rules are removed so that a long
	// --loop run does not exhaust them.
	mu     sync.Mutex
	nextUE uint32
	free   []uint32

	// qerMu guards the device's single QER, created by the first UE that needs it.
	// gtp5g refuses a QER that already exists, so creating one per UE would fail for
	// every UE after the first. A failed attempt is retried by the next UE.
	qerMu    sync.Mutex
	qerReady bool

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

// openClient prepares the device's long-lived netlink client.
func (l *sharedLink) openClient() {
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

	link, err := gtp5gnl.GetLink(l.name)
	if err != nil {
		conn.Close()
		mux.Close()
		log.Warn("[GNB][GTP] netlink client unavailable (link): ", err)

		return
	}

	l.mux = mux
	l.conn = conn
	l.client = client
	l.link = link
}

// PDRInstalled reports whether the rule really reached the datapath. It answers true
// when there is no client, so a device without one degrades to the old behaviour
// rather than declaring every UE broken.
func (l *sharedLink) PDRInstalled(id uint32) bool {
	l.clientMu.Lock()
	defer l.clientMu.Unlock()

	if l.client == nil || l.link == nil {
		return true
	}

	if _, err := gtp5gnl.GetPDR(l.client, l.link, int(id)); err != nil {
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
func (l *sharedLink) addRule(
	args []string,
	parse func([]string) ([]nl.Attr, error),
	create func(*gtp5gnl.Client, *gtp5gnl.Link, gtp5gnl.OID, []nl.Attr) error,
	fallback func([]string) error,
) error {
	l.clientMu.Lock()
	defer l.clientMu.Unlock()

	if l.client == nil || l.link == nil {
		return fallback(args)
	}

	if len(args) < 2 {
		return fmt.Errorf("too few parameters for a rule on %s", l.name)
	}

	oid, err := gtpTunnel.ParseOID(args[1])
	if err != nil {
		return err
	}

	attrs, err := parse(args[2:])
	if err != nil {
		return err
	}

	return create(l.client, l.link, oid, attrs)
}

// AddPDR installs a PDR through the device's own client.
func (l *sharedLink) AddPDR(args []string) error {
	return l.addRule(args, gtpTunnel.ParsePDROptions, gtp5gnl.CreatePDROID, gtpTunnel.CmdAddPDR)
}

// AddFAR installs a FAR through the device's own client.
func (l *sharedLink) AddFAR(args []string) error {
	return l.addRule(args, gtpTunnel.ParseFAROptions, gtp5gnl.CreateFAROID, gtpTunnel.CmdAddFAR)
}

// AddQER installs a QER through the device's own client.
func (l *sharedLink) AddQER(args []string) error {
	return l.addRule(args, gtpTunnel.ParseQEROptions, gtp5gnl.CreateQEROID, gtpTunnel.CmdAddQER)
}

// EnsureQER creates this device's QER the first time it is needed and reports whether
// it is usable, so callers can decide whether to reference it.
func (l *sharedLink) EnsureQER(create func() error) bool {
	l.qerMu.Lock()
	defer l.qerMu.Unlock()

	if l.qerReady {
		return true
	}

	if err := create(); err != nil {
		log.Error("[UE][GTP] Unable to create the shared QER: ", err)
		return false
	}

	l.qerReady = true

	return true
}

// ruleIDs are the rule identifiers one UE occupies on a device. PDR, FAR and QER are
// separate identifier spaces in gtp5g, so the same numbers may repeat across them.
type ruleIDs struct {
	slot    uint32
	pdrDown uint32
	pdrUp   uint32
	farDown uint32
	farUp   uint32
	qer     uint32
}

var (
	sharedLinksMu sync.Mutex
	sharedLinks   = make(map[string]*sharedLink)

	// setupSlots bounds how many tunnels are plumbed at once. Unbounded parallelism
	// fails in the netlink layer with "bad file descriptor"; the 500 ms registration
	// floor used to hide that by admitting at most two UEs a second.
	//
	// A single mutex is the wrong cure. SetupGtpInterface runs on the goroutine that
	// handles that UE's messages (see gnbMsgHandler), so serialising it backs up the
	// gNB's dispatcher: the simulator stops reading, the AMF's sends go unacked, and
	// the association eventually times out. That failure looks exactly like a core
	// problem and is not one. A small number of slots keeps the netlink layer inside
	// what it tolerates while leaving the dispatcher free to make progress.
	// Raised from 4 once rule creation stopped opening a socket per call: the bound
	// existed to contain that churn, not because the kernel objected to the work.
	// Override with PR_SETUP_SLOTS if a host needs it lower.
	setupSlots = make(chan struct{}, setupConcurrency())
)

// acquireSetupSlot blocks until this UE may plumb its tunnel.
func acquireSetupSlot() {
	setupSlots <- struct{}{}
}

func releaseSetupSlot() {
	<-setupSlots
}

// sharedLinkPrefix names every device this mode creates, so stale ones can be found.
const sharedLinkPrefix = "valgnb"

// sharedQERID is the single QER every UE on a shared device references.
const sharedQERID = 1

// routeTableOffset keeps per-UE routing tables clear of the identifiers Linux
// reserves: 253 (default), 254 (main) and 255 (local).
const routeTableOffset = 1000

// sharedLinkName derives a device name from the gNB's N3 address. Interface names are
// capped at 15 characters, and the hex form of an IPv4 address keeps this at 14 while
// staying unique per gNB.
func sharedLinkName(gnbIP netip.Addr) string {
	unmapped := gnbIP.Unmap()
	if !unmapped.Is4() {
		return sharedLinkPrefix + "shared"
	}

	v4 := unmapped.As4()

	return fmt.Sprintf("%s%02x%02x%02x%02x", sharedLinkPrefix, v4[0], v4[1], v4[2], v4[3])
}

// sharedLinkFor returns the device shared by every UE on this gNB, creating it on
// first use. Callers may arrive concurrently: UEs register in parallel, and only the
// first of them may create the device.
func sharedLinkFor(gnbIP netip.Addr) (*sharedLink, error) {
	name := sharedLinkName(gnbIP)

	sharedLinksMu.Lock()
	defer sharedLinksMu.Unlock()

	if link, ok := sharedLinks[name]; ok {
		return link, nil
	}

	// A device left behind by a previous run would still hold the socket.
	_ = gtpLink.CmdDel(name)

	link := &sharedLink{name: name, stop: make(chan bool)}

	go func() {
		// Does not return while the GTP-U socket is open, so it owns this goroutine
		// for the lifetime of the run rather than of one UE.
		if err := gtpLink.CmdAddWithStopCh(name, 1, 131072, gnbIP.String(), "", link.stop); err != nil {
			log.Error("[GNB][GTP] shared GTP device ", name, " ended: ", err)
		}
	}()

	if err := waitForLink(name, 5*time.Second); err != nil {
		close(link.stop)
		return nil, err
	}

	link.openClient()

	sharedLinks[name] = link
	log.Info(fmt.Sprintf("[GNB][GTP] shared GTP-U device %s created on %s; every UE of this gNB will use it", name, gnbIP))

	return link, nil
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

// take reserves one UE's worth of rule identifiers, reusing a slot a departed UE gave
// back before opening a new one. gtp5g's PDR identifier is 16 bits, so one device holds
// at most 32767 UEs' PDR pairs at once.
func (l *sharedLink) take() ruleIDs {
	l.mu.Lock()
	defer l.mu.Unlock()

	var n uint32
	if last := len(l.free) - 1; last >= 0 {
		n = l.free[last]
		l.free = l.free[:last]
	} else {
		n = l.nextUE
		l.nextUE++
	}

	return idsForSlot(n)
}

// giveBack makes a slot available to take again. Only call it once the slot's rules
// are gone from the device, or the next UE's creates collide with them.
func (l *sharedLink) giveBack(ids ruleIDs) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.free = append(l.free, ids.slot)
}

// release removes one UE's PDRs and FARs from the device and returns their slot. The
// QER is the device's, not the UE's, and stays. If a rule could not be removed the slot
// is not reused: its identifiers are still occupied.
func (l *sharedLink) release(ids ruleIDs) {
	if err := l.removeRules(ids); err != nil {
		log.Warn("[UE][GTP] Unable to remove the rules of slot ", ids.slot, " on ", l.name, "; not reusing it: ", err)
		return
	}

	l.giveBack(ids)
}

// removeRules deletes a UE's PDRs, then the FARs they reference. A rule that is already
// absent -- say, because setup failed before creating it -- counts as removed.
func (l *sharedLink) removeRules(ids ruleIDs) error {
	l.clientMu.Lock()
	defer l.clientMu.Unlock()

	type rule struct {
		id     uint32
		remove func(*gtp5gnl.Client, *gtp5gnl.Link, int) error
		cmd    func([]string) error
	}

	var errs []error
	for _, r := range []rule{
		{ids.pdrDown, gtp5gnl.RemovePDR, gtpTunnel.CmdDeletePDR},
		{ids.pdrUp, gtp5gnl.RemovePDR, gtpTunnel.CmdDeletePDR},
		{ids.farDown, gtp5gnl.RemoveFAR, gtpTunnel.CmdDeleteFAR},
		{ids.farUp, gtp5gnl.RemoveFAR, gtpTunnel.CmdDeleteFAR},
	} {
		var err error
		if l.client != nil && l.link != nil {
			err = r.remove(l.client, l.link, int(r.id))
		} else {
			err = r.cmd([]string{l.name, strconv.FormatUint(uint64(r.id), 10)})
		}

		if err != nil && !errors.Is(err, syscall.ENOENT) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// removeAddress takes one UE's address off a shared device.
func removeAddress(device string, ueIP string) {
	dev, err := netlink.LinkByName(device)
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
func idsForSlot(n uint32) ruleIDs {
	return ruleIDs{
		slot:    n,
		pdrDown: 2*n + 1,
		pdrUp:   2*n + 2,
		farDown: 2*n + 1,
		farUp:   2*n + 2,
		// Every UE of this gNB carries the same QFI, so one QER serves all of them
		// and PDRs simply point at it. Handing out a QER per UE also walked the
		// identifier past 255 at the 256th UE, which is where tunnel setup began
		// hanging inside the netlink layer.
		qer: sharedQERID,
	}
}

// dedicatedRuleIDs are the identifiers used when a UE owns its device outright, kept
// at their original values so that mode behaves exactly as before.
func dedicatedRuleIDs() ruleIDs {
	return ruleIDs{pdrDown: 1, pdrUp: 2, farDown: 1, farUp: 2, qer: sharedQERID}
}

// RemoveStaleSharedLink deletes the device this gNB will use if an earlier run left
// it behind. It has to happen before the gNB starts, not lazily at first tunnel setup:
// a stale device still owns the GTP-U socket on the N3 address, so the gNB cannot bind
// and the run dies before any UE gets there.
//
// Only this gNB's device is touched. Several PacketRusher processes can share a host,
// one per gNB, and removing every device named like ours would tear down the tunnels
// of the siblings that are already running.
func RemoveStaleSharedLink(gnbIP netip.Addr) {
	name := sharedLinkName(gnbIP)
	if _, err := netlink.LinkByName(name); err != nil {
		return // nothing left behind
	}

	if err := gtpLink.CmdDel(name); err != nil {
		log.Warn("[GNB][GTP] Unable to remove stale shared device ", name, ": ", err)
		return
	}

	log.Info("[GNB][GTP] Removed stale shared GTP-U device ", name, " from a previous run")
}

// setupConcurrency reports how many tunnels may be plumbed at once.
func setupConcurrency() int {
	if v := os.Getenv("PR_SETUP_SLOTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	return 32
}

// verifyRules reports whether each PDR should be read back after it is installed.
// Off by default: over 7236 UEs it found 0 absent rules and 0 retries, and it costs
// two netlink round trips per UE. Set PR_VERIFY_RULES=1 to turn it back on when
// diagnosing a datapath that is silently dropping.
func verifyRules() bool {
	return os.Getenv("PR_VERIFY_RULES") == "1"
}
