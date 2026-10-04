/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

// A UE's routing table is its own, and none of the kernel's.
func TestRoutingTable(t *testing.T) {
	require.Equal(t, 0x0a2d0002, routingTable(net.ParseIP("10.45.0.2").To4(), nil))
	require.Equal(t, 1<<23+254, routingTable(net.ParseIP("10.45.0.2").To4(), &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Index: 254}}))
	require.Equal(t, 1<<23+254, routingTable(nil, &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Index: 254}}))
}

// privilegedNetwork skips the test unless it may create devices, in a network
// namespace of its own, and gives lo the N3 addresses of the test's gNBs.
func privilegedNetwork(t *testing.T, gnbs ...string) {
	t.Helper()
	if os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("set PACKETRUSHER_TUN_TEST=1 in a private network namespace with CAP_NET_ADMIN")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, gnb := range gnbs {
		address, err := netlink.ParseAddr(gnb + "/32")
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, address))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, address)) })
	}
}

// message is what the gNB at the N3 address gnb sends its UE to set up the tunnel
// of its PDU session; device is the GTP-U device that gNB shares, if any.
func message(t *testing.T, gnb, upf string, teid uint32, device *gtp.Device) gnbContext.UEMessage {
	t.Helper()
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, upf, "01", "000000", 0, 9, 8, 9, teid, teid+1)
	require.NoError(t, err)
	return gnbContext.UEMessage{GNBPduSessions: [16]*gnbContext.GnbPDUSession{pdu}, GnbIp: netip.MustParseAddr(gnb), GtpDevice: device}
}

// fakeUPF listens for GTP-U on address until the test ends.
func fakeUPF(t *testing.T, address string) *net.UDPConn {
	t.Helper()
	upf, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(address), Port: 2152})
	require.NoError(t, err)
	t.Cleanup(func() { upf.Close() })
	return upf
}

// dialFromUE returns a UDP socket of an application using the UE's address family.
func dialFromUE(t *testing.T, session *context.UEPDUSession, local net.IP, remote string) net.Conn {
	t.Helper()
	dialer := net.Dialer{LocalAddr: &net.UDPAddr{IP: local}}
	if vrf := session.Tunnel.(*tunnel).vrf; vrf != nil {
		dialer.Control = func(_, _ string, c syscall.RawConn) (err error) {
			if controlErr := c.Control(func(fd uintptr) { err = syscall.BindToDevice(int(fd), vrf.Name) }); controlErr != nil {
				return controlErr
			}
			return err
		}
	}
	app, err := dialer.Dial("udp", remote)
	require.NoError(t, err)
	t.Cleanup(func() { app.Close() })
	return app
}

// echo sends a datagram from the UE's application and has the UPF return it: it
// must leave the gNB with the uplink TEID and come back on the downlink one.
func echo(t *testing.T, app net.Conn, upf *net.UDPConn, gnb string, uplink, downlink uint32) {
	t.Helper()
	echoThrough(t, app, upf, gnb, uplink, gnb, downlink)
}

// echoThrough is echo with the answer of the UPF sent to the gNB at the address answerTo.
func echoThrough(t *testing.T, app net.Conn, upf *net.UDPConn, gnb string, uplink uint32, answerTo string, downlink uint32) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	require.NoError(t, app.SetDeadline(deadline))
	require.NoError(t, upf.SetDeadline(deadline))
	ping := bytes.Repeat([]byte("ping"), 500) // more than an Ethernet frame carries
	_, err := app.Write(ping)
	require.NoError(t, err)
	buf := make([]byte, 65535)
	n, peer, err := upf.ReadFromUDPAddrPort(buf)
	require.NoError(t, err)
	require.Equal(t, netip.MustParseAddr(gnb), peer.Addr())
	teid, packet, err := userspace.Decode(buf[:n])
	require.NoError(t, err)
	require.Equal(t, uplink, teid)

	// Swapping the addresses and the ports keeps every checksum valid.
	addresses, ports, size := 12, 20, 4
	if packet[0]>>4 == 6 {
		addresses, ports, size = 8, 40, 16
	}
	reply := append(make([]byte, userspace.Headroom), packet...)
	inner := reply[userspace.Headroom:]
	copy(inner[addresses:], packet[addresses+size:addresses+2*size])
	copy(inner[addresses+size:], packet[addresses:addresses+size])
	copy(inner[ports:], packet[ports+2:ports+4])
	copy(inner[ports+2:], packet[ports:ports+2])
	// TS 29.281 §4.4.2 lets the UPF send from another port than 2152.
	other, err := net.DialUDP("udp4", &net.UDPAddr{IP: upf.LocalAddr().(*net.UDPAddr).IP}, &net.UDPAddr{IP: net.ParseIP(answerTo), Port: int(peer.Port())})
	require.NoError(t, err)
	defer other.Close()
	_, err = other.Write(userspace.Encode(reply, downlink, 9))
	require.NoError(t, err)
	n, err = app.Read(buf)
	require.NoError(t, err)
	require.Equal(t, ping, buf[:n])
}

// leftovers puts in the way of the UE what a killed run leaves on the host: devices
// with the names of the UE's, and the rule of its address, to a table that leads
// nowhere.
func leftovers(t *testing.T, ue *context.UEContext, session *context.UEPDUSession, devices bool) {
	t.Helper()
	if devices {
		for _, name := range []string{"gtp0", "gtp1", "val"} {
			require.NoError(t, netlink.LinkAdd(&netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name + ue.GetMsin()}}))
		}
		require.NoError(t, netlink.LinkAdd(&netlink.Vrf{LinkAttrs: netlink.LinkAttrs{Name: "vrf" + ue.GetMsin()}, Table: 4242}))
	}
	rule := netlink.NewRule()
	rule.Priority, rule.Table = 100, 4242
	rule.Src = &net.IPNet{IP: net.ParseIP(session.GetIp()).To4(), Mask: net.CIDRMask(32, 32)}
	require.NoError(t, netlink.RuleAdd(rule))
	blackhole := &netlink.Route{Dst: &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}, Table: 4242, Type: syscall.RTN_BLACKHOLE}
	require.NoError(t, netlink.RouteAdd(blackhole))
	t.Cleanup(func() { netlink.RouteDel(blackhole); netlink.RuleDel(rule) })
}

// released releases the tunnel of the session and checks that nothing is left of it.
func released(t *testing.T, ue *context.UEContext, session *context.UEPDUSession) {
	t.Helper()
	table := session.Tunnel.(*tunnel).table
	session.ReleaseTunnel()
	nothingLeft(t, ue, table)
}

// nothingLeft checks that the UE has no device, and no route or rule of its table.
func nothingLeft(t *testing.T, ue *context.UEContext, table int) {
	t.Helper()
	for _, name := range []string{"gtp0", "gtp1", "val", "vrf"} {
		_, err := netlink.LinkByName(name + ue.GetMsin())
		require.Error(t, err, name)
	}
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes)
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	require.NoError(t, err)
	for _, rule := range rules {
		require.NotEqual(t, 100, rule.Priority, "a rule of the UE is left: %s", rule)
	}
}

// One UE per backend and tunnel mode sends through its tunnel, is handed over to
// another gNB, hence another N3 address and TEIDs, then given another UPF there.
// The gNB at 127.88.4.3 cannot have a tunnel: its N3 address is taken.
func TestTunnelEndToEnd(t *testing.T) {
	privilegedNetwork(t, "127.88.4.1", "127.88.4.2")
	backends := []config.TunnelBackend{config.TunnelBackendUserspace}
	if _, err := os.Stat("/sys/module/gtp5g"); err == nil {
		backends = append(backends, config.TunnelBackendGtp5g)
	}
	if ebpfgtp.Load() == nil {
		t.Cleanup(ebpfgtp.Close)
		backends = append(backends, config.TunnelBackendEBPF)
	}
	modes := []config.TunnelMode{config.TunnelTun, config.TunnelVrf, config.TunnelShared}
	for number, backend := range backends {
		for _, mode := range modes {
			shared := mode == config.TunnelShared
			if shared && backend != config.TunnelBackendGtp5g {
				continue // Only gtp5g has a device that the UEs of a gNB share.
			}
			t.Run(fmt.Sprintf("%s/%s", backend, []string{"policy", "VRF", "shared"}[mode-config.TunnelTun]), func(t *testing.T) {
				upf, otherUPF, _ := fakeUPF(t, "127.88.4.8"), fakeUPF(t, "127.88.4.9"), fakeUPF(t, "127.88.4.3")
				var source, target *gtp.Device
				if shared {
					var err error
					source, err = gtp.NewDevice(netip.MustParseAddr("127.88.4.1"), 3000)
					require.NoError(t, err)
					t.Cleanup(source.Close)
					target, err = gtp.NewDevice(netip.MustParseAddr("127.88.4.2"), 3000)
					require.NoError(t, err)
					t.Cleanup(target.Close)
				}
				ue := &context.UEContext{TunnelMode: mode, TunnelBackend: backend, TunnelMTU: 3000}
				ue.UeSecurity.Msin = fmt.Sprintf("70055500%d%d", number, mode)
				session := &context.UEPDUSession{Id: 1}
				session.SetIp([12]uint8{10, 42, byte(number), byte(mode)})
				ue.PduSession[0] = session
				t.Cleanup(session.ReleaseTunnel)
				leftovers(t, ue, session, !shared)

				SetupGtpInterface(ue, message(t, "127.88.4.1", "127.88.4.8", 60, source))
				require.NotNil(t, session.Tunnel)
				tunnel := session.Tunnel.(*tunnel)
				first := tunnel.link.Attrs().Name
				link, err := netlink.LinkByName(first)
				require.NoError(t, err)
				require.Equal(t, 3000, link.Attrs().MTU)
				app := dialFromUE(t, session, net.ParseIP(session.GetIp()), "203.0.113.9:8888")
				echo(t, app, upf, "127.88.4.1", 60, 61)

				// A handover that fails leaves the UE the tunnel it has.
				SetupGtpInterface(ue, message(t, "127.88.4.3", "127.88.4.8", 90, nil))
				echo(t, app, upf, "127.88.4.1", 60, 61)

				// Handover: the application keeps its socket, and still receives what the
				// UPF sends to the source gNB until the core has switched the path.
				SetupGtpInterface(ue, message(t, "127.88.4.2", "127.88.4.8", 70, target))
				require.NotEqual(t, first, tunnel.link.Attrs().Name)
				echo(t, app, upf, "127.88.4.2", 70, 71)
				echoThrough(t, app, upf, "127.88.4.2", 70, "127.88.4.1", 61)
				tunnel.leave(tunnel.left)
				if !shared {
					_, err = netlink.LinkByName(first)
					require.Error(t, err, "the source gNB's device must be gone")
				}

				// Same gNB, another UPF and TEIDs: the device is updated in place.
				second := tunnel.link
				SetupGtpInterface(ue, message(t, "127.88.4.2", "127.88.4.9", 80, target))
				require.Same(t, second, tunnel.link)
				echo(t, app, otherUPF, "127.88.4.2", 80, 81)

				table := tunnel.table
				released(t, ue, session)
				if shared {
					require.Zero(t, source.WaitIdle(0, 0)+target.WaitIdle(0, 0), "the UE gave its rules back")
				} else {
					fakeUPF(t, "127.88.4.2").Close() // The N3 socket is released with its last UE.
				}

				// A first setup that fails leaves nothing behind.
				SetupGtpInterface(ue, message(t, "127.88.4.3", "127.88.4.8", 90, nil))
				require.Nil(t, session.Tunnel)
				nothingLeft(t, ue, table)
			})
		}
	}
}

// The UPF answers the UE's Router Solicitation with fe80::1 advertising the prefix
// 2001:db8:1:2::/64, in which the UE then has its address. A handover keeps it.
func TestTunnelIPv6(t *testing.T) {
	privilegedNetwork(t, "127.88.6.1", "127.88.6.2")
	advertisement, _ := hex.DecodeString("6000000000303aff" + "fe800000000000000000000000000001" +
		"ff020000000000000000000000000001" + "8600c386400007080000000000000000" +
		"030440c0ffffffffffffffff00000000" + "20010db8000100020000000000000000")
	for number, mode := range []config.TunnelMode{config.TunnelTun, config.TunnelVrf} {
		t.Run([]string{"IPv4v6 policy", "IPv6 VRF"}[number], func(t *testing.T) {
			upf := fakeUPF(t, "127.88.6.8")
			// A whole UE context, as the UE's goroutine is handed the advertised prefix.
			ue := &context.UEContext{TunnelBackend: config.TunnelBackendUserspace, TunnelMTU: 3000}
			ue.NewRanUeContext(fmt.Sprintf("700666000%d", number), &ie.UESecCapability{}, "", "", "", "", "", "001", "01",
				sidf.HomeNetworkPublicKey{}, "0000", "internet", 1, "", mode, nil, nil, 1)
			session := &context.UEPDUSession{Id: 1}
			ue.PduSession[0] = session
			t.Cleanup(session.ReleaseTunnel)
			if mode == config.TunnelTun {
				session.SetIp([12]uint8{10, 42, 6, 1})
			}
			session.SetIPv6(netip.MustParseAddr("fe80::7"))
			SetupGtpInterface(ue, message(t, "127.88.6.1", "127.88.6.8", 60, nil))
			require.NotNil(t, session.Tunnel)

			buf := make([]byte, 1500)
			require.NoError(t, upf.SetReadDeadline(time.Now().Add(3*time.Second)))
			n, gnb, err := upf.ReadFromUDPAddrPort(buf)
			require.NoError(t, err)
			teid, solicitation, err := userspace.Decode(buf[:n])
			require.NoError(t, err)
			require.Equal(t, uint32(60), teid)
			require.Equal(t, byte(133), solicitation[40])
			require.Equal(t, session.GetIPv6().AsSlice(), solicitation[8:24])
			_, err = upf.WriteToUDPAddrPort(userspace.Encode(append(make([]byte, userspace.Headroom), advertisement...), 61, 9), gnb)
			require.NoError(t, err)
			select {
			case plumb := <-ue.Deferred():
				plumb() // This test stands for the UE's goroutine.
			case <-time.After(3 * time.Second):
				t.Fatal("the advertised prefix was not handed to the UE")
			}
			address := netip.MustParseAddr("2001:db8:1:2::7")
			require.Equal(t, address, session.GetIPv6())
			app := dialFromUE(t, session, address.AsSlice(), "[2001:db8:ffff::9]:8888")
			echo(t, app, upf, "127.88.6.1", 60, 61)

			// Handover: the address follows to the new device, without a new solicitation.
			SetupGtpInterface(ue, message(t, "127.88.6.2", "127.88.6.8", 70, nil))
			require.Equal(t, "gtp1"+ue.GetMsin(), session.Tunnel.(*tunnel).link.Attrs().Name)
			echo(t, app, upf, "127.88.6.2", 70, 71)
			if session.GetIp() != "" {
				echo(t, dialFromUE(t, session, net.ParseIP(session.GetIp()), "203.0.113.9:8888"), upf, "127.88.6.2", 70, 71)
			}
			released(t, ue, session)
		})
	}
}
