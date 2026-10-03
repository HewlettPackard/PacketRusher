// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/binary"
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
	"golang.org/x/sys/unix"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

// udp6Checksum is an independent test-only checksum implementation over the
// IPv6 pseudo-header and UDP datagram, required for real host-stack replies.
func udp6Checksum(packet []byte) uint16 {
	data := append([]byte(nil), packet[8:40]...)
	data = binary.BigEndian.AppendUint32(data, uint32(len(packet)-40))
	data = append(data, 0, 0, 0, packet[6])
	data = append(data, packet[40:]...)
	if len(data)%2 != 0 {
		data = append(data, 0)
	}
	return ipChecksum(data)
}

func userspaceIPv6RoundTrip(t *testing.T, app net.Conn, peer *net.UDPConn, local netip.Addr, upID, downID uint32) {
	t.Helper()
	require.NoError(t, app.SetDeadline(time.Now().Add(3*time.Second)))
	require.NoError(t, peer.SetReadDeadline(time.Now().Add(3*time.Second)))
	_, err := app.Write([]byte("ping"))
	require.NoError(t, err)
	buffer := make([]byte, 65535)
	n, source, err := peer.ReadFromUDPAddrPort(buffer)
	require.NoError(t, err)
	require.Equal(t, local, source.Addr())
	teid, packet, err := userspace.Decode(buffer[:n])
	require.NoError(t, err)
	require.Equal(t, upID, teid)
	require.Equal(t, byte(6), packet[0]>>4)
	require.Equal(t, byte(17), packet[6])
	require.Equal(t, "ping", string(packet[48:]))
	reply := append([]byte(nil), packet...)
	copy(reply[8:24], packet[24:40])
	copy(reply[24:40], packet[8:24])
	copy(reply[40:42], packet[42:44])
	copy(reply[42:44], packet[40:42])
	copy(reply[48:], "pong")
	reply[46], reply[47] = 0, 0
	checksum := udp6Checksum(reply)
	if checksum == 0 {
		checksum = 65535
	}
	binary.BigEndian.PutUint16(reply[46:48], checksum)
	wire, err := userspace.Encode(downID, 9, reply)
	require.NoError(t, err)
	_, err = peer.WriteToUDPAddrPort(wire, source)
	require.NoError(t, err)
	n, err = app.Read(buffer)
	require.NoError(t, err)
	require.Equal(t, "pong", string(buffer[:n]))
}

func TestUserspaceRealIPv6DualStackPolicyVRFAndHandover(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("explicit isolated network namespace and CAP_NET_ADMIN required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, address := range []string{"127.88.4.1/32", "127.88.4.2/32"} {
		parsed, err := netlink.ParseAddr(address)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, parsed))
	}
	small := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "n3small", MTU: 1300}}
	require.NoError(t, netlink.LinkAdd(small))
	require.NoError(t, netlink.LinkSetUp(small))
	smallAddress, err := netlink.ParseAddr("127.88.4.3/32")
	require.NoError(t, err)
	require.NoError(t, netlink.AddrAdd(small, smallAddress))
	for _, family := range []uint8{ie.PDUSessType_IPv6, ie.PDUSessType_IPv4v6} {
		for _, mode := range []config.TunnelMode{config.TunnelTun, config.TunnelVrf} {
			t.Run(fmt.Sprintf("family%d-mode%d", family, mode), func(t *testing.T) {
				ue := &context.UEContext{TunnelMode: mode, TunnelBackend: config.TunnelBackendUserspace, TunnelMTU: 1456, PDUSessionType: config.PDUSessionType(family)}
				ue.UeSecurity.Msin = fmt.Sprintf("700666%02d%02d", family, mode)
				pdu, err := ue.CreatePDUSession()
				require.NoError(t, err)
				t.Cleanup(pdu.ReleaseTunnel)
				allocation := &ie.PDUAddr{IPv6IfId: []byte{0, 0, 0, 0, 0, 0, 0, 7}}
				if family == ie.PDUSessType_IPv4v6 {
					allocation.IPv4 = []byte{10, 42, family, byte(mode)}
				}
				require.NoError(t, pdu.SetPDUAddress(family, allocation))
				peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 9), Port: 2152})
				require.NoError(t, err)
				defer peer.Close()
				source := userspaceTestMessage(t, "127.88.4.1", 60)
				advertised := make(chan error, 1)
				go func() {
					buffer := make([]byte, 65535)
					_ = peer.SetReadDeadline(time.Now().Add(3 * time.Second))
					n, address, err := peer.ReadFromUDPAddrPort(buffer)
					if err != nil {
						advertised <- err
						return
					}
					teid, packet, err := userspace.Decode(buffer[:n])
					if err != nil {
						advertised <- err
						return
					}
					if teid != 60 || len(packet) != 48 || packet[40] != 133 || packet[7] != 255 {
						advertised <- fmt.Errorf("invalid RS through GTP-U")
						return
					}
					ra, _ := hex.DecodeString(raWire)
					invalid := append([]byte(nil), ra...)
					invalid[7] = 64
					wire, _ := userspace.Encode(61, 9, invalid)
					_, err = peer.WriteToUDPAddrPort(wire, address)
					if err != nil {
						advertised <- err
						return
					}
					time.Sleep(30 * time.Millisecond)
					if pdu.GetIPv6().IsValid() {
						advertised <- fmt.Errorf("untrusted RA installed a prefix")
						return
					}
					wire, _ = userspace.Encode(61, 9, ra)
					_, err = peer.WriteToUDPAddrPort(wire, address)
					advertised <- err
				}()
				SetupGtpInterface(ue, source)
				require.NoError(t, <-advertised)
				link := pdu.GetTunInterface()
				require.NotNil(t, link)
				require.Equal(t, "2001:db8:1234::7", pdu.GetIPv6().String())
				sessionRoutingTables.mu.Lock()
				table := int(sessionRoutingTables.sessions[pdu].table)
				sessionRoutingTables.mu.Unlock()
				routes, err := netlink.RouteListFiltered(netlink.FAMILY_V6, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
				require.NoError(t, err)
				var defaults []netlink.Route
				for _, route := range routes {
					if route.Dst == nil {
						defaults = append(defaults, route)
					} else if bits, _ := route.Dst.Mask.Size(); bits == 0 {
						defaults = append(defaults, route)
					}
				}
				require.Len(t, defaults, 1)
				require.Equal(t, link.Attrs().Index, defaults[0].LinkIndex)
				require.Equal(t, pdu.GetIPv6().String(), defaults[0].Src.String())
				mainRoutes, err := netlink.RouteListFiltered(netlink.FAMILY_V6, &netlink.Route{Table: unix.RT_TABLE_MAIN}, netlink.RT_FILTER_TABLE)
				require.NoError(t, err)
				for _, route := range mainRoutes {
					require.False(t, route.LinkIndex == link.Attrs().Index && route.Dst == nil, "RA must not create main-table default route")
				}
				dial := net.Dialer{LocalAddr: &net.UDPAddr{IP: net.IP(pdu.GetIPv6().AsSlice())}}
				if mode == config.TunnelVrf {
					name := pdu.GetVrfDevice().Attrs().Name
					dial.Control = func(_, _ string, connection syscall.RawConn) error {
						var setError error
						err := connection.Control(func(fd uintptr) {
							setError = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, name)
						})
						if err != nil {
							return err
						}
						return setError
					}
				}
				app, err := dial.Dial("udp6", "[2001:db8:ffff::9]:8888")
				require.NoError(t, err)
				defer app.Close()
				userspaceIPv6RoundTrip(t, app, peer, source.GnbIp, 60, 61)
				if family == ie.PDUSessType_IPv4v6 {
					v4dial := dial
					v4dial.LocalAddr = &net.UDPAddr{IP: net.ParseIP(pdu.GetIp())}
					v4app, err := v4dial.Dial("udp4", "203.0.113.9:8888")
					require.NoError(t, err)
					userspaceRoundTrip(t, v4app, peer, source.GnbIp, 60, 61)
					v4app.Close()
				}
				target := userspaceTestMessage(t, "127.88.4.2", 70)
				SetupGtpInterface(ue, target)
				require.Equal(t, link.Attrs().Index, pdu.GetTunInterface().Attrs().Index)
				userspaceIPv6RoundTrip(t, app, peer, target.GnbIp, 70, 71)
				// Automatic target-MTU calculation must fail before Linux disables
				// IPv6 on the live TUN and removes its negotiated addresses.
				ue.TunnelMTU = 0
				SetupGtpInterface(ue, userspaceTestMessage(t, "127.88.4.3", 80))
				require.Equal(t, target.GnbIp, pdu.GetGnbIp())
				require.Equal(t, 1456, link.Attrs().MTU)
				addresses, err := netlink.AddrList(link, netlink.FAMILY_V6)
				require.NoError(t, err)
				found := false
				for _, address := range addresses {
					if address.IP.String() == pdu.GetIPv6().String() {
						found = true
					}
				}
				require.True(t, found, "failed target setup must retain the source IPv6 address")
				userspaceIPv6RoundTrip(t, app, peer, target.GnbIp, 70, 71)
				app.Close()
				pdu.ReleaseTunnel()
				require.False(t, pdu.GetIPv6().IsValid())
				_, err = netlink.LinkByName(link.Attrs().Name)
				require.Error(t, err)
				routes, err = netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
				require.NoError(t, err)
				require.Empty(t, routes)
				rules, err := netlink.RuleList(netlink.FAMILY_ALL)
				require.NoError(t, err)
				for _, rule := range rules {
					require.NotEqual(t, table, rule.Table)
				}
			})
		}
	}
}
