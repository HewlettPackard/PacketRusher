// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

var ebpfRegistry = ebpfgtp.DefaultRegistry
var makeEBPFTUN = userspace.NewTUN
var disableEBPFEndpoint = netlink.LinkSetDown

type ebpfSession interface {
	Close() error
	Update(ebpfgtp.Config) error
	SetIPv6(netip.Addr) error
	SendControl([]byte) error
	SendUplink([]byte) error
	Advertisements() <-chan []byte
}

// Keep descriptors and routing reservations alive if kernel forwarding cannot
// be retired. Reusing an endpoint still named by a map would cross UE ownership.
var quarantinedEBPFEndpoints sync.Map

func (t *ebpfTunnel) quarantine(err error) {
	quarantinedEBPFEndpoints.Store(t, err)
	if t.table != nil {
		t.table.quarantine()
	}
	log.Error("[UE][eBPF] Quarantined endpoint after incomplete cleanup: ", err)
}

type ebpfTunnel struct {
	port        userspace.PacketPort
	session     ebpfSession
	link        netlink.Link
	rule        *netlink.Rule
	route       *netlink.Route
	vrf         *netlink.Vrf
	table       *routingTableReservation
	once        sync.Once
	ipv6Cleanup func() error
	uplinkDone  chan struct{}
}

func (t *ebpfTunnel) release() {
	t.once.Do(func() {
		if err := t.stopJumboUplink(); err != nil {
			var downErr error
			if t.link != nil {
				downErr = disableEBPFEndpoint(t.link)
			}
			t.quarantine(errors.Join(err, downErr))
			return
		}
		// Join allocation updates before removing their canonical authorization.
		removed := true
		if t.ipv6Cleanup != nil {
			removed = t.ipv6Cleanup() == nil
		}
		if t.session != nil {
			if err := t.session.Close(); err != nil {
				// Retaining ownership alone cannot stop an old canonical map.
				// Disable the owned endpoint before keeping its cleanup claim.
				var downErr error
				if t.link != nil {
					downErr = disableEBPFEndpoint(t.link)
				}
				t.quarantine(errors.Join(err, downErr))
				return
			}
		}
		// Removing the owned endpoint also removes any address whose deletion failed;
		// retain source policies until that operation has succeeded.
		if t.port != nil {
			if err := t.port.Close(); err != nil {
				t.quarantine(err)
				return
			}
		}
		if t.uplinkDone != nil {
			<-t.uplinkDone
		}
		if t.ipv6Cleanup != nil {
			removed = t.ipv6Cleanup() == nil
		}
		if t.route != nil {
			removed = routingObjectRemoved(routeDel(t.route)) && removed
		}
		if t.rule != nil {
			removed = routingObjectRemoved(ruleDel(t.rule)) && removed
		}
		if t.vrf != nil {
			removed = routingObjectRemoved(deleteTunnelLink(t.vrf)) && removed
		}
		if t.table != nil {
			if removed {
				t.table.release()
			} else {
				t.table.quarantine()
			}
		}
	})
}

func ebpfSessionConfig(pdu *context.UEPDUSession, gnbPDU *gnbContext.GnbPDUSession, local netip.Addr, endpoint, mtu int) (ebpfgtp.Config, error) {
	remote, err := netip.ParseAddr(gnbPDU.GetUpfIp())
	if err != nil {
		return ebpfgtp.Config{}, err
	}
	iid, ipv6 := pdu.GetIPv6InterfaceID()
	ip, err := netip.ParseAddr(pdu.GetIp())
	if err != nil && (pdu.GetIp() != "" || !ipv6) {
		return ebpfgtp.Config{}, err
	}
	if gnbPDU.GetQosId() < 0 || gnbPDU.GetQosId() > 63 {
		return ebpfgtp.Config{}, errors.New("eBPF QFI must be between 0 and 63")
	}
	cfg := ebpfgtp.Config{Local: local, Remote: remote, IPv4: ip, AllowIPv6: ipv6, IPv6InterfaceID: iid, IPv6: pdu.GetIPv6(), UplinkTEID: gnbPDU.GetTeidUplink(), DownlinkTEID: gnbPDU.GetTeidDownlink(), QFI: uint8(gnbPDU.GetQosId()), EndpointIfIndex: endpoint, MTU: mtu}
	return cfg, cfg.Validate()
}

// Reassembled outer IPv4 datagrams arrive at the owned UDP socket. Use a single
// nonblocking write while the file's RawConn protects descriptor ownership: this
// fallback must never keep the registry's joined receive worker stuck on a TUN.
func ebpfInjector(port userspace.PacketPort) (func([]byte) error, error) {
	conn, ok := port.(interface {
		SyscallConn() (syscall.RawConn, error)
	})
	if !ok {
		return nil, errors.New("eBPF TUN does not expose an owned nonblocking descriptor")
	}
	raw, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}
	return func(packet []byte) error {
		var writeErr error
		err := raw.Write(func(fd uintptr) bool {
			n, e := unix.Write(int(fd), packet)
			writeErr = e
			if e == nil && n != len(packet) {
				writeErr = io.ErrShortWrite
			}
			return true
		})
		return errors.Join(err, writeErr)
	}, nil
}

func (t *ebpfTunnel) setIPv6(address netip.Addr) error {
	if err := t.session.SetIPv6(address); err != nil {
		// Canonical update failure is fail-closed in the registry. Also disable the
		// owned endpoint before reporting failure if map deactivation was incomplete.
		downErr := disableEBPFEndpoint(t.link)
		stopErr := t.stopJumboUplink()
		t.quarantine(errors.Join(err, downErr, stopErr))
		return errors.Join(err, downErr, stopErr)
	}
	return nil
}

func setupEBPFTunnel(ue *context.UEContext, pdu *context.UEPDUSession, gnbPDU *gnbContext.GnbPDUSession, local netip.Addr) error {
	if pdu.GetTunInterface() != nil {
		return pdu.UpdateTunnel(gnbPDU, local)
	}
	t := &ebpfTunnel{}
	committed := false
	defer func() {
		if !committed {
			t.release()
		}
	}()
	var err error
	t.port, t.link, err = makeEBPFTUN(fmt.Sprintf("val%s", ue.GetMsin()))
	if err != nil {
		return err
	}
	_, ipv6 := pdu.GetIPv6InterfaceID()
	if err = setUserspaceTunnelMTU(t.link, local, ue.TunnelMTU, ipv6); err != nil {
		return err
	}
	if err = ebpfgtp.ConfigureEndpoint(t.link, t.port); err != nil {
		return err
	}
	cfg, err := ebpfSessionConfig(pdu, gnbPDU, local, t.link.Attrs().Index, t.link.Attrs().MTU)
	if err != nil {
		return err
	}
	cfg.Inject, err = ebpfInjector(t.port)
	if err != nil {
		return err
	}
	t.table, _, err = sessionRoutingTables.reserve(pdu)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: &net.IPNet{IP: net.IP(cfg.IPv4.AsSlice()), Mask: net.CIDRMask(32, 32)}}
	if ue.TunnelMode == config.TunnelVrf {
		t.vrf = &netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: fmt.Sprintf("vrf%s", ue.GetMsin())}, Table: t.table.table}
		if err = addTunnelVRF(t.vrf); err != nil {
			t.vrf = nil
			return err
		}
		if err = setTunnelMaster(t.link, t.vrf); err != nil {
			return err
		}
		if err = setTunnelUp(t.vrf); err != nil {
			return err
		}
	} else if cfg.IPv4.IsValid() {
		rule := netlink.NewRule()
		rule.Priority = 100
		rule.Table = int(t.table.table)
		rule.Src = addr.IPNet
		if err = addTunnelRule(rule); err != nil {
			return err
		}
		t.rule = rule
	}
	if cfg.IPv4.IsValid() {
		if err = addTunnelAddress(t.link, addr); err != nil {
			return err
		}
	}
	if err = setTunnelUp(t.link); err != nil {
		return err
	}
	ebpfRegistry.SetWarningHandler(func(err error) { log.Warn("[UE][eBPF] ", err) })
	// Publish a concrete nonnil result only after success (never a typed nil).
	session, err := ebpfRegistry.Open(cfg)
	if err != nil {
		return err
	}
	t.session = session
	if _, ok := t.port.(interface{ SetReadDeadline(time.Time) error }); !ok {
		return errors.New("owned eBPF TUN cannot cancel jumbo reads")
	}
	t.uplinkDone = make(chan struct{})
	go func() {
		defer close(t.uplinkDone)
		packet := make([]byte, 65535)
		for {
			n, err := t.port.Read(packet)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) && !errors.Is(err, os.ErrDeadlineExceeded) {
					log.Warn("[UE][eBPF] Owned jumbo uplink worker stopped: ", err)
				}
				return
			}
			if n <= 8000 {
				continue
			}
			if err = t.session.SendUplink(packet[:n]); err != nil && !errors.Is(err, net.ErrClosed) {
				log.Warn("[UE][eBPF] Jumbo uplink rejected: ", err)
			}
		}
	}()
	if cfg.IPv4.IsValid() {
		route := &netlink.Route{Dst: &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, LinkIndex: t.link.Attrs().Index, Scope: netlink.SCOPE_LINK, Protocol: 4, Priority: 1, Src: addr.IP, Table: int(t.table.table)}
		if err = replaceTunnelRoute(route); err != nil {
			return fmt.Errorf("install eBPF endpoint route: %w", err)
		}
		t.route = route
	}
	if cfg.AllowIPv6 {
		t.ipv6Cleanup, err = SetupIPv6Session(ue, pdu, t.link, t.table.table, t.session.SendControl, t.session.Advertisements(), t.setIPv6)
		if err != nil {
			return err
		}
	}
	pdu.SetTunInterface(t.link)
	pdu.SetUEInterface(t.link)
	pdu.SetTunRule(t.rule)
	pdu.SetTunRoute(t.route)
	pdu.SetVrfDevice(t.vrf)
	pdu.SetTunnelCleanup(func(bool) {
		t.release()
		pdu.SetTunInterface(nil)
		pdu.SetUEInterface(nil)
		pdu.SetTunRule(nil)
		pdu.SetTunRoute(nil)
		pdu.SetVrfDevice(nil)
	})
	pdu.SetTunnelUpdate(func(next *gnbContext.GnbPDUSession, ip netip.Addr) error {
		oldMTU := t.link.Attrs().MTU
		if err := setUserspaceTunnelMTU(t.link, ip, ue.TunnelMTU, ipv6); err != nil {
			return err
		}
		cfg, err := ebpfSessionConfig(pdu, next, ip, t.link.Attrs().Index, t.link.Attrs().MTU)
		if err == nil {
			cfg.Inject, err = ebpfInjector(t.port)
		}
		if err == nil {
			err = t.session.Update(cfg)
		}
		if err != nil {
			if restoreErr := restoreTunnelMTU(t.link, oldMTU); restoreErr != nil {
				return fmt.Errorf("%w: %v (restore MTU: %v)", errTunnelRollback, err, restoreErr)
			}
			return err
		}
		return nil
	})
	committed = true
	log.Infof("[UE][eBPF] TCX GTP-U configured on %s; IPv4 %s, IPv6 %s", t.link.Attrs().Name, pdu.GetIp(), pdu.GetIPv6())
	return nil
}

// Cancel a reader without closing/reusing an endpoint whose map retirement is
// still unproven. Its UDP send is bounded; completion precedes descriptor release.
func (t *ebpfTunnel) stopJumboUplink() error {
	if t.uplinkDone == nil {
		return nil
	}
	select {
	case <-t.uplinkDone:
		return nil
	default:
	}
	deadline := t.port.(interface{ SetReadDeadline(time.Time) error })
	if err := deadline.SetReadDeadline(time.Now()); err != nil {
		return err
	}
	<-t.uplinkDone
	return nil
}
