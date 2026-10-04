/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

package service

import (
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
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

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

func userspaceUE(t *testing.T, number int, mode config.TunnelMode) (*context.UEContext, *context.UEPDUSession) {
	t.Helper()
	ue, session := sharedSetupUE(t, number)
	ue.TunnelMode, ue.TunnelBackend, ue.TunnelMTU = mode, config.TunnelBackendUserspace, 1400
	return ue, session
}

func userspaceMessage(t *testing.T, gnb, upf string, teid uint32) gnbContext.UEMessage {
	t.Helper()
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, upf, "01", "000000", 0, 9, 8, 9, teid, teid+1)
	require.NoError(t, err)
	return gnbContext.UEMessage{GNBPduSessions: [16]*gnbContext.GnbPDUSession{pdu}, GnbIp: netip.MustParseAddr(gnb)}
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
	if vrf := session.GetVrfDevice(); vrf != nil {
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
	deadline := time.Now().Add(3 * time.Second)
	require.NoError(t, app.SetDeadline(deadline))
	require.NoError(t, upf.SetDeadline(deadline))
	_, err := app.Write([]byte("ping"))
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
	other, err := net.DialUDP("udp4", &net.UDPAddr{IP: upf.LocalAddr().(*net.UDPAddr).IP}, net.UDPAddrFromAddrPort(peer))
	require.NoError(t, err)
	defer other.Close()
	_, err = other.Write(userspace.Encode(reply, downlink, 9))
	require.NoError(t, err)
	n, err = app.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf[:n]))
}

// released releases the tunnel of the session and checks that nothing is left of it.
func released(t *testing.T, ue *context.UEContext, session *context.UEPDUSession) {
	t.Helper()
	table := session.GetTunRoute().Table
	session.ReleaseTunnel()
	for _, name := range []string{"gtp0", "gtp1", "val", "vrf"} {
		_, err := netlink.LinkByName(name + ue.GetMsin())
		require.Error(t, err, name)
	}
	used, err := routingTableInUse(uint32(table))
	require.NoError(t, err)
	require.False(t, used, "a route or a rule of the UE's routing table is left")
}

func TestUserspaceTunnelEndToEnd(t *testing.T) {
	privilegedNetwork(t, "127.88.4.1", "127.88.4.2")
	for number, mode := range []config.TunnelMode{config.TunnelTun, config.TunnelVrf} {
		t.Run([]string{"policy", "VRF"}[number], func(t *testing.T) {
			upf, otherUPF := fakeUPF(t, "127.88.4.8"), fakeUPF(t, "127.88.4.9")
			ue, session := userspaceUE(t, number+1, mode)
			SetupGtpInterface(ue, userspaceMessage(t, "127.88.4.1", "127.88.4.8", 60))
			require.NotNil(t, session.GetTunInterface())
			require.Equal(t, 1400, session.GetTunInterface().Attrs().MTU)
			app := dialFromUE(t, session, net.ParseIP(session.GetIp()), "203.0.113.9:8888")
			echo(t, app, upf, "127.88.4.1", 60, 61)

			// Handover: another gNB, hence another N3 address, device and TEIDs.
			SetupGtpInterface(ue, userspaceMessage(t, "127.88.4.2", "127.88.4.8", 70))
			require.Equal(t, "gtp1"+ue.GetMsin(), session.GetTunInterface().Attrs().Name)
			echo(t, app, upf, "127.88.4.2", 70, 71)
			_, err := netlink.LinkByName("gtp0" + ue.GetMsin())
			require.Error(t, err, "the source gNB's device must be gone")

			// Same gNB, another UPF and TEIDs: the device is updated in place.
			SetupGtpInterface(ue, userspaceMessage(t, "127.88.4.2", "127.88.4.9", 80))
			require.Equal(t, "gtp1"+ue.GetMsin(), session.GetTunInterface().Attrs().Name)
			echo(t, app, otherUPF, "127.88.4.2", 80, 81)

			released(t, ue, session)
			fakeUPF(t, "127.88.4.2").Close() // The N3 socket is released with its last UE.
		})
	}
}

// The UPF answers the UE's Router Solicitation with fe80::1 advertising the prefix
// 2001:db8:1:2::/64, in which the UE then has its address. A handover keeps it.
func TestUserspaceTunnelIPv6(t *testing.T) {
	privilegedNetwork(t, "127.88.6.1", "127.88.6.2")
	advertisement, _ := hex.DecodeString("6000000000303aff" + "fe800000000000000000000000000001" +
		"ff020000000000000000000000000001" + "8600c386400007080000000000000000" +
		"030440c0ffffffffffffffff00000000" + "20010db8000100020000000000000000")
	for number, mode := range []config.TunnelMode{config.TunnelTun, config.TunnelVrf} {
		t.Run([]string{"IPv4v6 policy", "IPv6 VRF"}[number], func(t *testing.T) {
			upf := fakeUPF(t, "127.88.6.8")
			// A whole UE context, as the UE's goroutine is handed the advertised prefix.
			ue := &context.UEContext{TunnelBackend: config.TunnelBackendUserspace, TunnelMTU: 1400}
			ue.NewRanUeContext(fmt.Sprintf("700666000%d", number), &ie.UESecCapability{}, "", "", "", "", "", "001", "01",
				sidf.HomeNetworkPublicKey{}, "0000", "internet", 1, "", mode, nil, nil, 1)
			session := &context.UEPDUSession{Id: 1}
			ue.PduSession[0] = session
			t.Cleanup(session.ReleaseTunnel)
			if mode == config.TunnelTun {
				session.SetIp([12]uint8{10, 42, 6, 1})
			}
			session.SetIPv6(netip.MustParseAddr("fe80::7"))
			SetupGtpInterface(ue, userspaceMessage(t, "127.88.6.1", "127.88.6.8", 60))
			require.NotNil(t, session.GetTunInterface())

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
			SetupGtpInterface(ue, userspaceMessage(t, "127.88.6.2", "127.88.6.8", 70))
			require.Equal(t, "gtp1"+ue.GetMsin(), session.GetTunInterface().Attrs().Name)
			echo(t, app, upf, "127.88.6.2", 70, 71)
			if session.GetIp() != "" {
				echo(t, dialFromUE(t, session, net.ParseIP(session.GetIp()), "203.0.113.9:8888"), upf, "127.88.6.2", 70, 71)
			}
			released(t, ue, session)
		})
	}
}
