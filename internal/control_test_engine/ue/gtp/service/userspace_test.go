// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestUserspaceConfigRejectsInvalidTEIDsAndQFI(t *testing.T) {
	msg := sharedSetupMessage(t, "127.88.4.1", 10)
	cfg, err := userspaceConfig(msg.GNBPduSessions[0], msg.GnbIp, "10.0.0.1")
	require.NoError(t, err)
	require.Equal(t, uint8(0), cfg.QFI)
	_, err = userspaceConfig(msg.GNBPduSessions[0], msg.GnbIp, "not-an-address")
	require.Error(t, err)
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, "127.88.0.9", "01", "000000", 0, 64, 8, 9, 10, 11)
	require.NoError(t, err)
	_, err = userspaceConfig(pdu, msg.GnbIp, "10.0.0.1")
	require.Error(t, err)
}

func TestUserspaceUpdateFailureReleasesOnlyWhenRollbackFails(t *testing.T) {
	for _, rollbackFails := range []bool{false, true} {
		t.Run(fmt.Sprintf("rollbackFails=%t", rollbackFails), func(t *testing.T) {
			ue, pdu := sharedSetupUE(t, 1)
			ue.TunnelBackend = config.TunnelBackendUserspace
			source := userspaceTestMessage(t, "127.88.4.1", 60)
			target := userspaceTestMessage(t, "127.88.4.2", 70)
			link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Name: "existing", MTU: 1456}}
			pdu.SetTunInterface(link)
			pdu.SetGnbIp(source.GnbIp)
			pdu.GnbPduSession = source.GNBPduSessions[0]
			released := 0
			pdu.SetTunnelCleanup(func(bool) { released++; pdu.SetTunInterface(nil) })
			pdu.SetTunnelUpdate(func(next *gnbContext.GnbPDUSession, ip netip.Addr) error {
				require.Same(t, target.GNBPduSessions[0], next)
				require.Equal(t, target.GnbIp, ip)
				if rollbackFails {
					return fmt.Errorf("%w: restore source MTU failed", errTunnelRollback)
				}
				return errors.New("target bind failed; source restored")
			})
			SetupGtpInterface(ue, target)
			require.Equal(t, source.GnbIp, pdu.GetGnbIp())
			require.Same(t, source.GNBPduSessions[0], pdu.GnbPduSession)
			if rollbackFails {
				require.Equal(t, 1, released)
				require.Nil(t, pdu.GetTunInterface())
			} else {
				require.Zero(t, released)
				require.Same(t, link, pdu.GetTunInterface())
			}
			pdu.ReleaseTunnel()
			require.Equal(t, 1, released, "cleanup must run once")
		})
	}
}
func userspaceTestMessage(t *testing.T, local string, teid uint32) gnbContext.UEMessage {
	t.Helper()
	gnbUE := &gnbContext.GNBUe{}
	gnbUE.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
	pdu, err := gnbUE.CreatePduSession(1, "127.88.4.9", "01", "000000", 0, 9, 8, 9, teid, teid+1)
	require.NoError(t, err)
	var sessions [16]*gnbContext.GnbPDUSession
	sessions[0] = pdu
	return gnbContext.UEMessage{GNBPduSessions: sessions, GnbIp: netip.MustParseAddr(local)}
}
func ipChecksum(b []byte) uint16 {
	var sum uint32
	for len(b) > 1 {
		sum += uint32(binary.BigEndian.Uint16(b))
		b = b[2:]
	}
	for sum>>16 != 0 {
		sum = (sum & 65535) + (sum >> 16)
	}
	return ^uint16(sum)
}
func userspaceRoundTrip(t *testing.T, app net.Conn, peer *net.UDPConn, local netip.Addr, upID, downID uint32) {
	t.Helper()
	require.NoError(t, app.SetDeadline(time.Now().Add(3*time.Second)))
	require.NoError(t, peer.SetReadDeadline(time.Now().Add(3*time.Second)))
	_, err := app.Write([]byte("ping"))
	require.NoError(t, err)
	buf := make([]byte, 65535)
	n, source, err := peer.ReadFromUDPAddrPort(buf)
	require.NoError(t, err)
	require.Equal(t, local, source.Addr())
	id, ip, err := userspace.Decode(buf[:n])
	require.NoError(t, err)
	require.Equal(t, upID, id)
	reply := append([]byte(nil), ip...)
	ihl := int(reply[0]&15) * 4
	copy(reply[12:16], ip[16:20])
	copy(reply[16:20], ip[12:16])
	reply[10], reply[11] = 0, 0
	binary.BigEndian.PutUint16(reply[10:12], ipChecksum(reply[:ihl]))
	copy(reply[ihl:ihl+2], ip[ihl+2:ihl+4])
	copy(reply[ihl+2:ihl+4], ip[ihl:ihl+2])
	reply[ihl+6], reply[ihl+7] = 0, 0
	copy(reply[ihl+8:], "pong")
	wire, err := userspace.Encode(downID, 9, reply)
	require.NoError(t, err)
	_, err = peer.WriteToUDPAddrPort(wire, source)
	require.NoError(t, err)
	n, err = app.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "pong", string(buf[:n]))
}
func TestUserspaceRealPolicyVRFHandoverAndRollback(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("explicit isolated network namespace and CAP_NET_ADMIN required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, host := range []string{"127.88.4.1/32", "127.88.4.2/32", "127.88.4.3/32"} {
		address, err := netlink.ParseAddr(host)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, address))
	}
	for _, mode := range []config.TunnelMode{config.TunnelShared, config.TunnelVrf} {
		t.Run(map[config.TunnelMode]string{config.TunnelShared: "policy", config.TunnelVrf: "VRF"}[mode], func(t *testing.T) {
			ue, pdu := sharedSetupUE(t, int(mode))
			ue.TunnelMode = mode
			ue.TunnelBackend = config.TunnelBackendUserspace
			ue.TunnelMTU = 1456
			peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 9), Port: 2152})
			require.NoError(t, err)
			defer peer.Close()
			source := userspaceTestMessage(t, "127.88.4.1", 60)
			SetupGtpInterface(ue, source)
			link := pdu.GetTunInterface()
			require.NotNil(t, link)
			table := pdu.GetTunRoute().Table
			require.GreaterOrEqual(t, table, firstRoutingTable)
			require.Equal(t, 1456, link.Attrs().MTU)
			dial := net.Dialer{LocalAddr: &net.UDPAddr{IP: net.ParseIP(pdu.GetIp())}}
			if mode == config.TunnelVrf {
				name := pdu.GetVrfDevice().Attrs().Name
				dial.Control = func(_, _ string, c syscall.RawConn) error {
					var setErr error
					err := c.Control(func(fd uintptr) { setErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, name) })
					if err != nil {
						return err
					}
					return setErr
				}
			}
			app, err := dial.Dial("udp4", "203.0.113.9:8888")
			require.NoError(t, err)
			defer app.Close()
			userspaceRoundTrip(t, app, peer, source.GnbIp, 60, 61)
			target := userspaceTestMessage(t, "127.88.4.2", 70)
			SetupGtpInterface(ue, target)
			require.Equal(t, link.Attrs().Index, pdu.GetTunInterface().Attrs().Index)
			require.Equal(t, table, pdu.GetTunRoute().Table)
			userspaceRoundTrip(t, app, peer, target.GnbIp, 70, 71)
			blocked, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 3), Port: 2152})
			require.NoError(t, err)
			defer blocked.Close()
			SetupGtpInterface(ue, userspaceTestMessage(t, "127.88.4.3", 80))
			require.Equal(t, target.GnbIp, pdu.GetGnbIp())
			require.Equal(t, target.GNBPduSessions[0], pdu.GnbPduSession)
			userspaceRoundTrip(t, app, peer, target.GnbIp, 70, 71)
			blocked.Close()
			// A target bind failure normally preserves the source. If restoring the
			// already-changed TUN MTU also fails, retire that inconsistent tunnel.
			blocked, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 3), Port: 2152})
			require.NoError(t, err)
			defer blocked.Close()
			previousRestore := restoreTunnelMTU
			t.Cleanup(func() { restoreTunnelMTU = previousRestore })
			restores := 0
			restoreTunnelMTU = func(actual netlink.Link, previous int) error {
				restores++
				require.Equal(t, link.Attrs().Index, actual.Attrs().Index)
				require.Equal(t, 1456, previous)
				require.Equal(t, 1400, actual.Attrs().MTU, "target MTU must be applied before its bind fails")
				kernelLink, err := netlink.LinkByName(actual.Attrs().Name)
				require.NoError(t, err)
				require.Equal(t, 1400, kernelLink.Attrs().MTU)
				return errors.New("injected source MTU restoration failure")
			}
			ue.TunnelMTU = 1400
			SetupGtpInterface(ue, userspaceTestMessage(t, "127.88.4.3", 80))
			restoreTunnelMTU = previousRestore
			require.Equal(t, 1, restores)
			require.Equal(t, target.GnbIp, pdu.GetGnbIp())
			require.Same(t, target.GNBPduSessions[0], pdu.GnbPduSession)
			require.Nil(t, pdu.GetTunInterface())
			app.Close()
			pdu.ReleaseTunnel()
			require.Nil(t, pdu.GetTunInterface())
			_, err = netlink.LinkByName(link.Attrs().Name)
			require.Error(t, err)
			routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
			require.NoError(t, err)
			require.Empty(t, routes)
			rules, err := netlink.RuleList(netlink.FAMILY_ALL)
			require.NoError(t, err)
			for _, rule := range rules {
				require.NotEqual(t, table, rule.Table)
			}
			rebound, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(target.GnbIp, 2152)))
			require.NoError(t, err)
			rebound.Close()
		})
	}
}
