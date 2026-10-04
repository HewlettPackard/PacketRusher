/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"

	gtpLink "github.com/free5gc/go-gtp5gnl/linkcmd"
	gtpTunnel "github.com/free5gc/go-gtp5gnl/tuncmd"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// gtp5gCommands create, or modify, the rules of a gtp5g device from the arguments
// of gtp5g's own commands.
type gtp5gCommands struct{ far, pdr, qer func([]string) error }

var (
	createRules = gtp5gCommands{gtpTunnel.CmdAddFAR, gtpTunnel.CmdAddPDR, gtpTunnel.CmdAddQER}
	modifyRules = gtp5gCommands{gtpTunnel.CmdModFAR, gtpTunnel.CmdModPDR, gtpTunnel.CmdModQER}
)

// run applies the rules of one UE to the gtp5g device name, under the identifiers
// ids: a FAR and a PDR per direction. Its uplink is marked with the QFI of the QER
// qer, if not 0; c.qer is nil when that QER is not the UE's own to create.
func (c gtp5gCommands) run(name string, ids gtp.RuleIDs, qer uint32, pdu *gnbContext.GnbPDUSession, ueIP string, gnbIP netip.Addr) error {
	id := func(n uint32) string { return strconv.FormatUint(uint64(n), 10) }
	type rule struct {
		command func([]string) error
		args    []string
	}
	rules := []rule{
		// Both directions forward (action 2), the uplink under the GTP-U/UDP/IPv4
		// header (0) of the UPF's F-TEID.
		{c.far, []string{name, id(ids.FARDown), "--action", "2"}},
		{c.far, []string{name, id(ids.FARUp), "--action", "2", "--hdr-creation", "0", id(pdu.GetTeidUplink()), pdu.GetUpfIp(), "2152"}},
		// Downlink, from the core (interface 1): what the UPF sends to the gNB's
		// F-TEID, without its GTP-U/UDP/IPv4 header (0).
		{c.pdr, []string{name, id(ids.PDRDown), "--pcd", "1", "--hdr-rm", "0", "--ue-ipv4", ueIP,
			"--f-teid", id(pdu.GetTeidDownlink()), gnbIP.String(), "--far-id", id(ids.FARDown), "--src-intf", "1"}},
	}
	// Uplink, from the access (interface 0): what the UE sends from its address.
	// The GTP-U source address is gtp5g's own, not part of the PFCP specification.
	uplink := []string{name, id(ids.PDRUp), "--pcd", "2", "--ue-ipv4", ueIP, "--far-id", id(ids.FARUp), "--src-intf", "0",
		"--gtpu-src-ip", gnbIP.String()}
	if qer != 0 {
		uplink = append(uplink, "--qer-id", id(qer))
		if c.qer != nil {
			rules = append(rules, rule{c.qer, []string{name, id(qer), "--qfi", strconv.FormatInt(pdu.GetQosId(), 10)}})
		}
	}
	for _, r := range append(rules, rule{c.pdr, uplink}) {
		log.Debug("[UE][GTP] GTP rule ", strings.Join(r.args, " "))
		if err := r.command(r.args); err != nil {
			return fmt.Errorf("GTP rule %s: %w", strings.Join(r.args, " "), err)
		}
	}
	return nil
}

// gtp5gLink is the gtp5g datapath of a UE that has a device of its own.
type gtp5gLink struct {
	name string
	stop chan bool
	done chan struct{} // closed once the goroutine holding the GTP-U socket has returned
}

// dedicatedQER is the UE's own QER on its device, if its flow has a QFI to mark.
func dedicatedQER(pdu *gnbContext.GnbPDUSession) uint32 {
	if pdu.GetQosId() > 0 {
		return gtp.DedicatedQERID
	}
	return 0
}

func startGtp5g(name string, pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) (datapath, error) {
	if ueIP == "" {
		return nil, errNoIPv6
	}
	t := &gtp5gLink{name: name, stop: make(chan bool), done: make(chan struct{})}
	ended := make(chan error, 1)
	go func() {
		defer close(t.done)
		// Returns once stopped, or at once if it cannot bind the N3 address.
		ended <- gtpLink.CmdAddWithStopCh(name, 1, 131072, ip.String(), "", t.stop)
	}()
	for start := time.Now(); ; time.Sleep(20 * time.Millisecond) {
		if _, err := netlink.LinkByName(name); err == nil {
			break
		}
		select {
		case err := <-ended:
			return nil, fmt.Errorf("GTP device %s: %w", name, err)
		default:
		}
		if time.Since(start) > 5*time.Second {
			t.close()
			return nil, fmt.Errorf("GTP device %s did not appear within 5s", name)
		}
	}
	if err := createRules.run(name, gtp.DedicatedRuleIDs(), dedicatedQER(pdu), pdu, ueIP, ip); err != nil {
		t.close()
		return nil, err
	}
	return t, nil
}

// refresh modifies the rules in place: the device and its socket stay.
func (t *gtp5gLink) refresh(pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) error {
	return modifyRules.run(t.name, gtp.DedicatedRuleIDs(), dedicatedQER(pdu), pdu, ueIP, ip)
}

func (t *gtp5gLink) close() {
	close(t.stop)
	// The goroutine owns the GTP-U socket, even once the device is gone: no other
	// session can bind the N3 address before it has returned.
	<-t.done
	removeLink(t.name)
}

func (t *gtp5gLink) solicit(netip.Addr) (netip.Addr, error) {
	return netip.Addr{}, errNoIPv6
}

// sharedRules is the gtp5g datapath of a UE on the device its gNB shares among its
// UEs: the UE's rules on that device, which also holds its address.
type sharedRules struct {
	dev  *gtp.Device
	ids  gtp.RuleIDs
	ueIP string
}

func startShared(dev *gtp.Device, pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) (datapath, error) {
	if dev == nil {
		return nil, errors.New("the gNB has no shared GTP-U device")
	}
	s := &sharedRules{dev: dev, ueIP: ueIP}
	var err error
	if s.ids, err = s.addRules(pdu, ip); err != nil {
		return nil, err
	}
	link, err := netlink.LinkByName(dev.Name())
	if err == nil {
		err = netlink.AddrAdd(link, &netlink.Addr{IPNet: &net.IPNet{IP: net.ParseIP(ueIP).To4(), Mask: net.CIDRMask(32, 32)}})
	}
	if err != nil {
		dev.Release(s.ids)
		return nil, fmt.Errorf("UE address on %s: %w", dev.Name(), err)
	}
	return s, nil
}

// addRules takes identifiers on the device and creates the UE's rules under them.
func (s *sharedRules) addRules(pdu *gnbContext.GnbPDUSession, ip netip.Addr) (gtp.RuleIDs, error) {
	ids, err := s.dev.Take()
	if err != nil {
		return ids, err
	}
	// QERs belong to the device, one per QFI, not to the UE: gtp5g refuses a QER
	// that already exists.
	var qer uint32
	if qfi := pdu.GetQosId(); qfi > 0 {
		id, usable := s.dev.EnsureQER(qfi, func(id uint32) error {
			return s.dev.AddQER([]string{s.dev.Name(), strconv.FormatUint(uint64(id), 10), "--qfi", strconv.FormatInt(qfi, 10)})
		})
		if usable {
			qer = id
		}
	}
	// Through the device's long-lived netlink client, rather than one per command.
	err = gtp5gCommands{far: s.dev.AddFAR, pdr: s.dev.AddPDR}.run(s.dev.Name(), ids, qer, pdu, s.ueIP, ip)
	if err != nil {
		s.dev.Release(ids)
	}
	return ids, err
}

// refresh replaces the UE's rules: the device stays, and the UE's address on it.
func (s *sharedRules) refresh(pdu *gnbContext.GnbPDUSession, _ string, ip netip.Addr) error {
	ids, err := s.addRules(pdu, ip)
	if err == nil {
		s.dev.Release(s.ids)
		s.ids = ids
	}
	return err
}

// close gives back what the UE holds of the device, which is the gNB's to remove.
func (s *sharedRules) close() {
	s.dev.RemoveAddress(s.ueIP)
	s.dev.Release(s.ids)
}

func (s *sharedRules) solicit(netip.Addr) (netip.Addr, error) {
	return netip.Addr{}, errNoIPv6
}
