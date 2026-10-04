// SPDX-License-Identifier: Apache-2.0
package service

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"io"
	"my5G-RANTester/config"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// The independent UPF fixture emits the downlink PDU session information type.
// Encode's default type is uplink; userspace accepts both, kernel admission is
// deliberately strict about the direction received from its configured UPF.
func downlinkGTP(teid uint32, qfi uint8, payload []byte) ([]byte, error) {
	wire, err := userspace.Encode(teid, qfi, payload)
	if err == nil && qfi != 0 {
		wire[13] = 0
	}
	return wire, err
}

func TestEBPFAllocationFailureDisablesOwnedEndpoint(t *testing.T) {
	failure := errors.New("canonical prefix update failed")
	tunnel := &ebpfTunnel{session: &ebpfFailureSession{err: failure}, link: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 99}}}
	previous := disableEBPFEndpoint
	calls := 0
	disableEBPFEndpoint = func(link netlink.Link) error { require.Equal(t, 99, link.Attrs().Index); calls++; return nil }
	t.Cleanup(func() { disableEBPFEndpoint = previous; quarantinedEBPFEndpoints.Delete(tunnel) })
	require.ErrorIs(t, tunnel.setIPv6(netip.MustParseAddr("2001:db8::7")), failure)
	require.Equal(t, 1, calls)
	_, retained := quarantinedEBPFEndpoints.Load(tunnel)
	require.True(t, retained)
}

func TestEBPFParityTCPPeerProcess(t *testing.T) { testpeer.RunTCP(t) }

func TestNativeEBPFProductionTCPFamiliesPolicyVRFAndDeviceBinding(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private privileged namespace required")
	}
	runProductionTCPFamiliesPolicyVRFAndDeviceBinding(t, config.TunnelBackendEBPF)
}
func TestNativeUserspaceProductionTCPFamiliesPolicyVRFAndDeviceBinding(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" && os.Getenv("PACKETRUSHER_TUN_TEST") != "1" {
		t.Skip("private privileged namespace required")
	}
	runProductionTCPFamiliesPolicyVRFAndDeviceBinding(t, config.TunnelBackendUserspace)
}
func TestNativeProductionIPv6TCPVirtualOffloadAndDeviceBinding(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	for _, backend := range []config.TunnelBackend{config.TunnelBackendEBPF, config.TunnelBackendUserspace} {
		t.Run(string(backend), func(t *testing.T) { runProductionTCPFamiliesPolicyVRFAndDeviceBinding(t, backend, true) })
	}
}
func runProductionTCPFamiliesPolicyVRFAndDeviceBinding(t *testing.T, backend config.TunnelBackend, keepOffload ...bool) {
	previous := ebpfRegistry
	ebpfRegistry = ebpfgtp.NewRegistry()
	defer func() { ebpfRegistry = previous }()
	families, modes := []int{4, 6}, []config.TunnelMode{config.TunnelTun, config.TunnelVrf}
	virtual := len(keepOffload) != 0 && keepOffload[0]
	if virtual {
		families, modes = []int{6}, []config.TunnelMode{config.TunnelVrf}
		t.Setenv("PACKETRUSHER_TCP_KEEP_OFFLOAD", "1")
	}
	for _, family := range families {
		for _, mode := range modes {
			t.Run(fmt.Sprintf("IPv%d-mode%d", family, mode), func(t *testing.T) {
				t.Setenv("PACKETRUSHER_TCP_FAMILY", strconv.Itoa(family))
				join := testpeer.Start(t, "TestEBPFParityTCPPeerProcess")
				if !virtual {
					out, err := exec.Command("ethtool", "-K", "pr-n3", "tx", "off", "rx", "off", "gro", "off", "gso", "off", "tso", "off").CombinedOutput()
					require.NoError(t, err, string(out))
				} else {
					out, err := exec.Command("ethtool", "-k", "pr-n3").CombinedOutput()
					require.NoError(t, err, string(out))
					t.Logf("unchanged virtual N3 offloads: %s", out)
				}
				ue, pdu := sharedSetupUE(t, 1)
				if family == 6 {
					ue, pdu = ipv6TestSession(t)
					ue.UeSecurity.Msin = "7005550001"
				}
				ue.TunnelBackend = backend
				ue.TunnelMode = mode
				if family == 4 {
					pdu.SetIp([12]uint8{10, 60, 0, 1})
				}
				destination, network := "192.0.2.1:9000", "tcp4"
				if family == 6 {
					ue.PDUSessionType = config.PDUSessionType(ie.PDUSessType_IPv6)
					require.NoError(t, pdu.SetPDUAddress(ie.PDUSessType_IPv6, &ie.PDUAddr{IPv6IfId: []byte{0, 0, 0, 0, 0, 0, 0, 7}}))
					destination, network = "[2001:db8:ffff::9]:9000", "tcp6"
				}
				source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
				if backend == config.TunnelBackendEBPF {
					require.NoError(t, setupEBPFTunnel(ue, pdu, source.GNBPduSessions[0], source.GnbIp))
				} else {
					require.NoError(t, setupUserspaceTunnel(ue, pdu, source.GNBPduSessions[0], source.GnbIp))
				}
				t.Cleanup(pdu.ReleaseTunnel)
				device := pdu.GetTunInterface()
				require.NotNil(t, device)
				ip := net.ParseIP(pdu.GetIp())
				if family == 6 {
					ip = net.IP(pdu.GetIPv6().AsSlice())
				}
				app, err := (&net.Dialer{Timeout: 5 * time.Second, LocalAddr: &net.TCPAddr{IP: ip}, Control: func(_, _ string, raw syscall.RawConn) error {
					var setErr error
					err := raw.Control(func(fd uintptr) {
						setErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, device.Attrs().Name)
					})
					return errors.Join(err, setErr)
				}}).Dial(network, destination)
				require.NoError(t, err)
				defer app.Close()
				require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
				payload := testpeer.TCPBulkPayload()
				_, err = io.CopyN(app, bytes.NewReader(payload), int64(len(payload)))
				require.NoError(t, err)
				reply := make([]byte, len(payload))
				_, err = io.ReadFull(app, reply)
				require.NoError(t, err)
				require.Equal(t, payload, reply)
				_, err = app.Write([]byte{1})
				require.NoError(t, err)
				require.NoError(t, app.Close())
				join()
				if backend == config.TunnelBackendEBPF {
					ul, dl, dropped, err := ebpfRegistry.Stats()
					require.NoError(t, err)
					require.Greater(t, ul, uint64(2))
					if virtual {
						require.Zero(t, dl, "unchanged virtual RX partial checksum is validated by owned UDP fallback")
					} else {
						require.Greater(t, dl, uint64(2))
					}
					t.Logf("SO_BINDTODEVICE=%s family%d mode%d TCP exact bytes%d kernel UL%d DL%d drop%d", device.Attrs().Name, family, mode, len(payload), ul, dl, dropped)
				} else {
					t.Logf("userspace SO_BINDTODEVICE=%s family%d mode%d TCP exact bytes%d", device.Attrs().Name, family, mode, len(payload))
				}

				pdu.ReleaseTunnel()
				require.Empty(t, sessionRoutingTables.sessions)
				require.Nil(t, pdu.GetTunInterface())
			})
		}
	}
}

func TestEBPFRetirementDisableFailureRetainsBothErrors(t *testing.T) {
	closeErr, downErr := errors.New("canonical deletion and tombstone failed"), errors.New("endpoint disable failed")
	session := &ebpfFailureSession{err: closeErr}
	port := &ebpfFailurePort{}
	tunnel := &ebpfTunnel{session: session, port: port, link: &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 99}}}
	previous := disableEBPFEndpoint
	disabled := 0
	disableEBPFEndpoint = func(netlink.Link) error { disabled++; return downErr }
	t.Cleanup(func() { disableEBPFEndpoint = previous; quarantinedEBPFEndpoints.Delete(tunnel) })
	tunnel.release()
	tunnel.release()
	cause, retained := quarantinedEBPFEndpoints.Load(tunnel)
	require.True(t, retained)
	require.ErrorIs(t, cause.(error), closeErr)
	require.ErrorIs(t, cause.(error), downErr)
	require.Equal(t, 1, disabled)
	require.Zero(t, port.closed, "retain the exact owned descriptor while map and endpoint deactivation are unproven")
}

func TestNativeEBPFVirtualN3ChecksumFallbackAndGatewayMark(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	for _, profile := range []string{"virtual-offload", "gateway-mark"} {
		t.Run(profile, func(t *testing.T) {
			previous := ebpfRegistry
			ebpfRegistry = ebpfgtp.NewRegistry()
			defer func() { ebpfRegistry = previous }()
			step := "1"
			remote := "10.88.0.2"
			if profile == "virtual-offload" {
				t.Setenv("PACKETRUSHER_PEER_KEEP_OFFLOAD", "1")
			} else {
				remote = "10.99.0.2"
				step = "1-remote"
				t.Setenv("PACKETRUSHER_EBPF_REMOTE_ADDR", remote)
			}
			join := testpeer.Start(t, "TestEBPFServicePeerProcess", step)
			if profile == "gateway-mark" {
				n3, err := netlink.LinkByName("pr-n3")
				require.NoError(t, err)
				route := &netlink.Route{Dst: testpeer.Network(remote + "/32"), Gw: net.ParseIP("10.88.0.2"), LinkIndex: n3.Attrs().Index}
				require.NoError(t, netlink.RouteAdd(route))
				defer netlink.RouteDel(route)
				trap := &netlink.Route{Dst: testpeer.Network("0.0.0.0/0"), Table: 9999, Type: unix.RTN_BLACKHOLE}
				require.NoError(t, netlink.RouteAdd(trap))
				defer netlink.RouteDel(trap)
				rule := netlink.NewRule()
				rule.Priority = 110
				rule.Mark = 0x42
				rule.Table = 9999
				require.NoError(t, netlink.RuleAdd(rule))
				defer netlink.RuleDel(rule)
			}
			ue, pdu := sharedSetupUE(t, 1)
			ue.TunnelBackend = config.TunnelBackendEBPF
			pdu.SetIp([12]uint8{10, 60, 0, 1})
			source := ebpfTestMessage(t, "10.88.0.1", 1001, 2001)
			source.GNBPduSessions[0].SetUpfIp(remote)
			require.NoError(t, setupEBPFTunnel(ue, pdu, source.GNBPduSessions[0], source.GnbIp))
			t.Cleanup(pdu.ReleaseTunnel)
			device := pdu.GetTunInterface()
			app, err := (&net.Dialer{LocalAddr: &net.UDPAddr{IP: net.ParseIP(pdu.GetIp())}, Control: func(_, _ string, raw syscall.RawConn) error {
				var setErr error
				err := raw.Control(func(fd uintptr) {
					setErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, device.Attrs().Name)
					if setErr == nil && profile == "gateway-mark" {
						setErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, 0x42)
					}
				})
				return errors.Join(err, setErr)
			}}).Dial("udp4", "192.0.2.1:9000")
			require.NoError(t, err)
			defer app.Close()
			require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
			_, err = app.Write([]byte(profile))
			require.NoError(t, err)
			buf := make([]byte, 100)
			n, err := app.Read(buf)
			require.NoError(t, err)
			require.Equal(t, profile, string(buf[:n]))
			join()
			ul, dl, drops, err := ebpfRegistry.Stats()
			require.NoError(t, err)
			require.Equal(t, uint64(1), ul)
			if profile == "virtual-offload" {
				require.Zero(t, dl, "valid virtual partial checksum uses owned UDP validation and injector")
			} else {
				require.Equal(t, uint64(1), dl)
				require.GreaterOrEqual(t, drops, uint64(2))
			}
			t.Logf("actual %s UL%d kernelDL%d kernelDrops%d, bound interface%s", profile, ul, dl, drops, device.Attrs().Name)
			pdu.ReleaseTunnel()
			require.Empty(t, sessionRoutingTables.sessions)
		})
	}
}

func TestNativeEBPFJumboOptionsAndZeroQFI(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		a, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, a))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
	}
	previous := ebpfRegistry
	ebpfRegistry = ebpfgtp.NewRegistry()
	defer func() { ebpfRegistry = previous }()
	for _, qfi := range []uint8{0, 9} {
		t.Run(fmt.Sprintf("QFI%d", qfi), func(t *testing.T) {
			for _, payloadLength := range []int{7001, 65459} {
				t.Run(fmt.Sprintf("payload%d", payloadLength), func(t *testing.T) {
					ue, pdu := sharedSetupUE(t, 1)
					ue.TunnelBackend = config.TunnelBackendEBPF
					ue.TunnelMTU = 8000
					if payloadLength > 8000 {
						ue.TunnelMTU = 0
					}
					pdu.SetIp([12]uint8{10, 60, 0, 1})
					gu := &gnbContext.GNBUe{}
					gu.CreateUeContext("not informed", "", []string{"01"}, []string{"000000"}, nil)
					gp, err := gu.CreatePduSession(1, "127.88.4.9", "01", "000000", 0, int64(qfi), 8, 9, 60, 61)
					require.NoError(t, err)
					peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 9), Port: 2152})
					require.NoError(t, err)
					defer peer.Close()
					require.NoError(t, setupEBPFTunnel(ue, pdu, gp, netip.MustParseAddr("127.88.4.1")))
					defer pdu.ReleaseTunnel()
					device := pdu.GetTunInterface()
					expectedMTU := 8000
					if payloadLength > 8000 {
						expectedMTU = 65491
					}
					require.Equal(t, expectedMTU, device.Attrs().MTU)
					app, err := (&net.Dialer{LocalAddr: &net.UDPAddr{IP: net.ParseIP(pdu.GetIp())}, Control: func(_, _ string, raw syscall.RawConn) error {
						var setErr error
						err := raw.Control(func(fd uintptr) {
							setErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, device.Attrs().Name)
							if setErr == nil {
								setErr = unix.SetsockoptString(int(fd), unix.IPPROTO_IP, unix.IP_OPTIONS, string([]byte{1, 1, 1, 0}))
							}
						})
						return errors.Join(err, setErr)
					}}).Dial("udp4", "192.0.2.1:9000")
					require.NoError(t, err)
					defer app.Close()
					require.NoError(t, app.SetDeadline(time.Now().Add(time.Second)))
					require.NoError(t, peer.SetDeadline(time.Now().Add(time.Second)))
					payload := bytes.Repeat([]byte{0x71}, payloadLength)
					_, err = app.Write(payload)
					require.NoError(t, err)
					buf := make([]byte, 65535)
					n, address, err := peer.ReadFromUDPAddrPort(buf)
					require.NoError(t, err)
					expectedFlag := byte(0x34)
					if qfi == 0 {
						expectedFlag = 0x30
					}
					require.Equal(t, expectedFlag, buf[0])
					teid, inner, err := userspace.Decode(buf[:n])
					require.NoError(t, err)
					require.Equal(t, uint32(60), teid)
					ihl := int(inner[0]&15) * 4
					require.Equal(t, 24, ihl)
					require.Zero(t, ipChecksum(inner[:ihl]))
					require.Equal(t, payload, inner[ihl+8:])
					if binary.BigEndian.Uint16(inner[ihl+6:ihl+8]) != 0 {
						pseudo := append([]byte(nil), inner[12:20]...)
						pseudo = append(pseudo, 0, 17, inner[ihl+4], inner[ihl+5])
						pseudo = append(pseudo, inner[ihl:]...)
						require.Zero(t, ipChecksum(pseudo), "jumbo uplink checksum is complete at UPF")
					}
					reply := append([]byte(nil), inner...)
					copy(reply[12:16], inner[16:20])
					copy(reply[16:20], inner[12:16])
					copy(reply[ihl:ihl+2], inner[ihl+2:ihl+4])
					copy(reply[ihl+2:ihl+4], inner[ihl:ihl+2])
					reply[10], reply[11], reply[ihl+6], reply[ihl+7] = 0, 0, 0, 0
					binary.BigEndian.PutUint16(reply[10:12], ipChecksum(reply[:ihl]))
					wire, err := downlinkGTP(61, qfi, reply)
					require.NoError(t, err)
					_, err = peer.WriteToUDPAddrPort(wire, address)
					require.NoError(t, err)
					n, err = app.Read(buf)
					require.NoError(t, err)
					require.Equal(t, payload, buf[:n])
					require.NoError(t, app.Close())
					pdu.ReleaseTunnel()
					require.Empty(t, sessionRoutingTables.sessions)
					require.Nil(t, pdu.GetTunInterface())
					socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 88, 4, 1), Port: 2152})
					require.NoError(t, err)
					require.NoError(t, socket.Close())
					t.Logf("actual jumbo%d-byte UDP+IPv4options+QFI%d+boundTUN roundtrip and joined reader/table/port reuse", payloadLength, qfi)
				})
			}

		})
	}
}

// Retain the real active canonical map while injecting unavailable update and
// retirement primitives. The service must disable its actual owned endpoint;
// reserving the objects alone would leave this same downlink forwarding.
type nativeFailedEBPFSession struct {
	ebpfSession
	failure error
}

func (s *nativeFailedEBPFSession) SetIPv6(netip.Addr) error { return s.failure }
func (s *nativeFailedEBPFSession) Close() error             { return s.failure }

func TestNativeEBPFFailedAuthorizationDisablesTrafficAndRetainsClaims(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("private namespace required")
	}
	lo, err := netlink.LinkByName("lo")
	require.NoError(t, err)
	for _, cidr := range []string{"127.88.4.1/32", "127.88.4.9/32"} {
		a, err := netlink.ParseAddr(cidr)
		require.NoError(t, err)
		require.NoError(t, netlink.AddrAdd(lo, a))
		t.Cleanup(func() { require.NoError(t, netlink.AddrDel(lo, a)) })
	}
	for _, operation := range []string{"prefix-update", "canonical-retirement"} {
		t.Run(operation, func(t *testing.T) {
			_, pdu := sharedSetupUE(t, 1)
			port, link, err := userspace.NewTUN("pr-fail-closed")
			require.NoError(t, err)
			defer port.Close()
			require.NoError(t, netlink.LinkSetMTU(link, 1456))
			require.NoError(t, netlink.LinkSetUp(link))
			require.NoError(t, netlink.AddrAdd(link, &netlink.Addr{IPNet: testpeer.Network("10.60.0.1/32")}))
			require.NoError(t, netlink.RouteAdd(&netlink.Route{Dst: testpeer.Network("192.0.2.1/32"), LinkIndex: link.Attrs().Index, Scope: netlink.SCOPE_LINK}))
			inject, err := ebpfInjector(port)
			require.NoError(t, err)
			cfg := ebpfgtp.Config{Local: netip.MustParseAddr("127.88.4.1"), Remote: netip.MustParseAddr("127.88.4.9"), IPv4: netip.MustParseAddr("10.60.0.1"), AllowIPv6: true, IPv6InterfaceID: [8]byte{0, 0, 0, 0, 0, 0, 0, 7}, IPv6: netip.MustParseAddr("2001:db8:1234::7"), UplinkTEID: 60, DownlinkTEID: 61, QFI: 9, EndpointIfIndex: link.Attrs().Index, MTU: 1456, Inject: inject}
			r := ebpfgtp.NewRegistry()
			session, err := r.Open(cfg)
			require.NoError(t, err)
			defer session.Close()
			table, _, err := sessionRoutingTables.reserve(pdu)
			require.NoError(t, err)
			defer table.release()
			failure := errors.New("injected BPF mutation and deactivation failure")
			tunnel := &ebpfTunnel{session: &nativeFailedEBPFSession{ebpfSession: session, failure: failure}, port: port, link: link, table: table}
			defer quarantinedEBPFEndpoints.Delete(tunnel)
			peer, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(cfg.Remote, 2152)))
			require.NoError(t, err)
			defer peer.Close()
			app, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IP(cfg.IPv4.AsSlice()), Port: 50000})
			require.NoError(t, err)
			defer app.Close()
			inner := make([]byte, 32)
			inner[0], inner[8], inner[9] = 0x45, 64, 17
			binary.BigEndian.PutUint16(inner[2:4], uint16(len(inner)))
			copy(inner[12:16], []byte{192, 0, 2, 1})
			copy(inner[16:20], cfg.IPv4.AsSlice())
			binary.BigEndian.PutUint16(inner[10:12], ipChecksum(inner[:20]))
			binary.BigEndian.PutUint16(inner[20:22], 9000)
			binary.BigEndian.PutUint16(inner[22:24], 50000)
			binary.BigEndian.PutUint16(inner[24:26], 12)
			copy(inner[28:], "live")
			wire, err := downlinkGTP(61, 9, inner)
			require.NoError(t, err)
			send := func() {
				_, err := peer.WriteToUDP(wire, net.UDPAddrFromAddrPort(netip.AddrPortFrom(cfg.Local, 2152)))
				require.NoError(t, err)
			}
			send()
			require.NoError(t, app.SetReadDeadline(time.Now().Add(time.Second)))
			buf := make([]byte, 32)
			n, _, err := app.ReadFromUDP(buf)
			require.NoError(t, err)
			require.Equal(t, "live", string(buf[:n]))
			if operation == "prefix-update" {
				require.ErrorIs(t, tunnel.setIPv6(cfg.IPv6), failure)
			} else {
				tunnel.release()
			}
			current, err := netlink.LinkByIndex(link.Attrs().Index)
			require.NoError(t, err)
			require.Zero(t, current.Attrs().Flags&net.FlagUp)
			cause, retained := quarantinedEBPFEndpoints.Load(tunnel)
			require.True(t, retained)
			require.ErrorIs(t, cause.(error), failure)
			routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: int(table.table)}, netlink.RT_FILTER_TABLE)
			require.NoError(t, err)
			require.Len(t, routes, 1)
			require.Equal(t, int(table.table), routes[0].Table)
			require.Equal(t, int(unix.RTN_BLACKHOLE), routes[0].Type)
			require.Equal(t, table.claim.Priority, routes[0].Priority)
			require.NotContains(t, sessionRoutingTables.free, table.table)
			fresh, _, err := sessionRoutingTables.reserve(pdu)
			require.NoError(t, err)
			require.NotEqual(t, table.table, fresh.table)
			fresh.release()
			defer func() { require.NoError(t, netlink.RouteDel(table.claim)) }()
			send()
			require.NoError(t, app.SetReadDeadline(time.Now().Add(30*time.Millisecond)))
			_, _, err = app.ReadFromUDP(buf)
			require.True(t, os.IsTimeout(err), "retained canonical map must not deliver through the disabled owned endpoint")
			t.Logf("actual %s: initial downlink delivered; TUN DOWN, later traffic blocked, canonical/port/table claims retained", operation)
		})
	}
}
