//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"bytes"
	"io"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/userspace"
)

func TestEBPFTCPPeerProcess(t *testing.T) { testpeer.RunTCP(t) }

func TestNativeTCPBulkTrafficAndCleanup(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	join := testpeer.Start(t, "TestEBPFTCPPeerProcess")
	// A physical NIC completes inner partial checksums before the UPF sees the
	// wire. Veth transports offload metadata instead; disable it on owned N3.
	out, err := exec.Command("ethtool", "-K", "pr-n3", "tx", "off", "rx", "off", "gro", "off", "gso", "off", "tso", "off").CombinedOutput()
	require.NoError(t, err, string(out))
	endpoint, device, err := userspace.NewTUN("pr-ue")
	require.NoError(t, err)
	defer endpoint.Close()
	require.NoError(t, netlink.LinkSetMTU(device, 1456))
	require.NoError(t, netlink.AddrAdd(device, &netlink.Addr{IPNet: testpeer.Network("10.60.0.1/32")}))
	require.NoError(t, ConfigureEndpoint(device))
	require.NoError(t, netlink.LinkSetUp(device))
	r := NewRegistry()
	cfg := Config{Local: netip.MustParseAddr("10.88.0.1"), Remote: netip.MustParseAddr("10.88.0.2"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 1001, DownlinkTEID: 2001, QFI: 9, EndpointIfIndex: device.Attrs().Index, MTU: 1456}
	session, err := r.Open(cfg)
	require.NoError(t, err)
	defer session.Close()
	rule := netlink.NewRule()
	rule.Priority, rule.Table, rule.Src = 100, 1600, testpeer.Network("10.60.0.1/32")
	require.NoError(t, netlink.RuleAdd(rule))
	defer netlink.RuleDel(rule)
	encap := &netlink.BpfEncap{}
	require.NoError(t, encap.SetProg(nl.LWT_BPF_XMIT, session.ProgramFD(), "packetrusher_gtpu"))
	require.NoError(t, encap.SetXmitHeadroom(44))
	route := &netlink.Route{Dst: testpeer.Network("0.0.0.0/0"), LinkIndex: device.Attrs().Index, Scope: netlink.SCOPE_LINK, Table: 1600, Src: net.ParseIP("10.60.0.1"), Encap: encap}
	require.NoError(t, netlink.RouteAdd(route))
	defer netlink.RouteDel(route)
	app, err := (&net.Dialer{Timeout: 5 * time.Second, LocalAddr: &net.TCPAddr{IP: net.ParseIP("10.60.0.1")}}).Dial("tcp4", "192.0.2.1:9000")
	if err != nil {
		ul, dl, dropped, statsErr := r.Stats()
		t.Logf("failed TCP handshake kernel UL=%d DL=%d ingressDrops=%d statsErr=%v", ul, dl, dropped, statsErr)
	}
	require.NoError(t, err)
	defer app.Close()
	t.Log("actual UE-bound TCP handshake completed; beginning bulk transfer")
	require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
	payload := testpeer.TCPBulkPayload()
	written, err := io.CopyN(app, bytes.NewReader(payload), int64(len(payload)))
	ul, dl, dropped, statsErr := r.Stats()
	t.Logf("bulk write bytes=%d err=%v kernel UL=%d DL=%d ingressDrops=%d statsErr=%v", written, err, ul, dl, dropped, statsErr)
	require.NoError(t, err)
	reply := make([]byte, len(payload))
	_, err = io.ReadFull(app, reply)
	require.NoError(t, err)
	require.Equal(t, payload, reply)
	_, err = app.Write([]byte{1})
	require.NoError(t, err)
	ul, dl, _, err = r.Stats()
	require.NoError(t, err)
	require.Greater(t, ul, uint64(2), "actual uplink must contain handshake and bulk segments")
	require.Greater(t, dl, uint64(2), "actual downlink must contain handshake and bulk segments")
	t.Logf("actual kernel TCP bulk bytes=%d uplink=%d downlink=%d", len(payload), ul, dl)
	require.NoError(t, app.Close())
	join()
	require.NoError(t, netlink.RouteDel(route))
	require.NoError(t, session.Close())
	require.Empty(t, r.sessions)
	require.Empty(t, r.links)
	require.Empty(t, r.locals)
	require.Nil(t, r.collection)
}
