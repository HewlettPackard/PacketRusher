// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	gtpTunnel "github.com/free5gc/go-gtp5gnl/tuncmd"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"strconv"
	"sync"
	"time"

	gtpLink "github.com/free5gc/go-gtp5gnl/linkcmd"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	modifyTunnelPDR     = gtpTunnel.CmdModPDR
	modifyTunnelFAR     = gtpTunnel.CmdModFAR
	modifyTunnelQER     = gtpTunnel.CmdModQER
	addDedicatedLink    = gtpLink.CmdAddWithStopCh
	deleteDedicatedLink = gtpLink.CmdDel
	findTunnelLink      = netlink.LinkByName
	deleteTunnelLink    = netlink.LinkDel
)

type dedicatedTunnel struct {
	name          string
	stop          chan bool
	done          chan struct{}
	stopOnce      sync.Once
	endpoint      netlink.Link
	endpointOwned bool
	vrf           *netlink.Vrf
	vrfOwned      bool
	rule          *netlink.Rule
	route         *netlink.Route
	table         *routingTableReservation
	activeRules   dedicatedRules
}

// The goroutine owns a duplicate of the UDP socket even after LinkDel. Waiting
// for it to return is necessary before another session can bind the N3 address.
func startDedicatedTunnel(name string, ip netip.Addr) (*dedicatedTunnel, error) {
	if _, err := findTunnelLink(name); err == nil {
		if err := deleteDedicatedLink(name); err != nil {
			return nil, err
		}
	}
	tunnel := &dedicatedTunnel{name: name, stop: make(chan bool), done: make(chan struct{})}
	ended := make(chan error, 1)
	go func() {
		defer close(tunnel.done)
		ended <- addDedicatedLink(name, 1, 131072, ip.String(), "", tunnel.stop)
	}()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case err := <-ended:
			if err == nil {
				err = fmt.Errorf("GTP socket worker ended before %s was ready", name)
			}
			tunnel.release(false)
			return nil, err
		case <-deadline.C:
			tunnel.release(false)
			return nil, fmt.Errorf("GTP device %s did not appear within 5s", name)
		case <-tick.C:
			if _, err := findTunnelLink(name); err == nil {
				return tunnel, nil
			}
		}
	}
}

func (t *dedicatedTunnel) release(retiring bool) {
	routingRemoved := true
	if !retiring {
		if t.route != nil {
			routingRemoved = routingObjectRemoved(routeDel(t.route))
		}
		if t.rule != nil {
			routingRemoved = routingObjectRemoved(ruleDel(t.rule)) && routingRemoved
		}
	}
	t.stopOnce.Do(func() { close(t.stop) })
	<-t.done // No new netlink sockets until the worker's mux has finished closing.
	if err := deleteDedicatedLink(t.name); err != nil {
		log.Warn("[UE][GTP] Unable to remove GTP backend ", t.name, ": ", err)
	}
	if !retiring {
		if t.endpointOwned && t.endpoint != nil {
			_ = deleteTunnelLink(t.endpoint)
		}
		if t.vrfOwned && t.vrf != nil {
			routingRemoved = routingObjectRemoved(deleteTunnelLink(t.vrf)) && routingRemoved
		}
		if routingRemoved {
			t.table.release()
		} else {
			t.table.quarantine()
			log.Warn("[UE][GTP] Keeping routing table reservation after incomplete cleanup for ", t.name)
		}
	}
}

func createUEEndpoint(name string) (netlink.Link, error) {
	endpoint := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name}}
	if err := netlink.LinkAdd(endpoint); err != nil {
		return nil, err
	}
	link, err := netlink.LinkByName(name)
	if err != nil {
		_ = netlink.LinkDel(endpoint)
		return nil, err
	}
	if err := netlink.LinkSetUp(link); err != nil {
		_ = netlink.LinkDel(link)
		return nil, err
	}
	return link, nil
}

var errTunnelRollback = errors.New("GTP rule rollback failed")

type dedicatedRules struct {
	farUp, pdrDown, pdrUp, qer []string
}

func dedicatedRuleSet(name string, pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) dedicatedRules {
	rules := dedicatedRules{
		farUp:   []string{name, "2", "--action", "2", "--hdr-creation", "0", strconv.FormatUint(uint64(pdu.GetTeidUplink()), 10), pdu.GetUpfIp(), "2152"},
		pdrDown: []string{name, "1", "--pcd", "1", "--hdr-rm", "0", "--ue-ipv4", ueIP, "--f-teid", strconv.FormatUint(uint64(pdu.GetTeidDownlink()), 10), ip.String(), "--far-id", "1", "--src-intf", "1"},
		pdrUp:   []string{name, "2", "--pcd", "2", "--ue-ipv4", ueIP, "--far-id", "2", "--src-intf", "0", "--gtpu-src-ip", ip.String()},
	}
	if qfi := pdu.GetQosId(); qfi > 0 {
		rules.qer = []string{name, "1", "--qfi", strconv.FormatInt(qfi, 10)}
		rules.pdrUp = append(rules.pdrUp, "--qer-id", "1")
	}
	return rules
}

func (r dedicatedRules) operations() []struct {
	args   []string
	modify func([]string) error
} {
	operations := []struct {
		args   []string
		modify func([]string) error
	}{
		{r.farUp, modifyTunnelFAR}, {r.pdrDown, modifyTunnelPDR},
	}
	if r.qer != nil {
		operations = append(operations, struct {
			args   []string
			modify func([]string) error
		}{r.qer, modifyTunnelQER})
	}
	return append(operations, struct {
		args   []string
		modify func([]string) error
	}{r.pdrUp, modifyTunnelPDR})
}

// Same-N3 updates retain the socket and host network objects. Failed updates
// replay every previous rule because an error can follow a partially applied call.
func (t *dedicatedTunnel) refresh(pdu *gnbContext.GnbPDUSession, ueIP string, ip netip.Addr) error {
	next := dedicatedRuleSet(t.name, pdu, ueIP, ip)
	previous := t.activeRules
	if (next.qer == nil) != (previous.qer == nil) {
		return errors.New("changing QFI enabled/disabled on an existing N3 tunnel requires a new PDU session")
	}
	for _, operation := range next.operations() {
		if err := operation.modify(operation.args); err != nil {
			var rollback []error
			for _, restore := range previous.operations() {
				if err := restore.modify(restore.args); err != nil {
					rollback = append(rollback, err)
				}
			}
			if len(rollback) != 0 {
				return errors.Join(err, errTunnelRollback, errors.Join(rollback...))
			}
			return fmt.Errorf("update failed and previous GTP rules restored: %w", err)
		}
	}
	t.activeRules = next
	return nil
}
