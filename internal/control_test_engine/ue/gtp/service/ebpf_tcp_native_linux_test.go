//go:build linux

// SPDX-License-Identifier: Apache-2.0
package service

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/ebpfgtp"
	"my5G-RANTester/internal/control_test_engine/ue/gtp/internal/testpeer"
)

func TestEBPFServiceTCPPeerProcess(t *testing.T) { testpeer.RunTCP(t) }

func TestNativeEBPFServiceTCPBulkTrafficAndRelease(t *testing.T) {
	if os.Getenv("PACKETRUSHER_EBPF_TEST") != "1" {
		t.Skip("requires private privileged namespace")
	}
	join := testpeer.Start(t, "TestEBPFServiceTCPPeerProcess")
	out, err := exec.Command("ethtool", "-K", "pr-n3", "tx", "off", "rx", "off", "gro", "off", "gso", "off", "tso", "off").CombinedOutput()
	require.NoError(t, err, string(out))
	previous := ebpfRegistry
	ebpfRegistry = ebpfgtp.NewRegistry()
	t.Cleanup(func() { ebpfRegistry = previous })
	ue, pdu := sharedSetupUE(t, 1)
	ue.TunnelBackend = config.TunnelBackendEBPF
	pdu.SetIp([12]uint8{10, 60, 0, 1})
	SetupGtpInterface(ue, ebpfTestMessage(t, "10.88.0.1", 1001, 2001))
	require.NotNil(t, pdu.GetTunInterface())
	defer pdu.ReleaseTunnel()
	device := pdu.GetTunInterface()
	table := pdu.GetTunRoute().Table
	app, err := (&net.Dialer{Timeout: 5 * time.Second, LocalAddr: &net.TCPAddr{IP: net.ParseIP("10.60.0.1")}}).Dial("tcp4", "192.0.2.1:9000")
	require.NoError(t, err)
	defer app.Close()
	require.NoError(t, app.SetDeadline(time.Now().Add(5*time.Second)))
	payload := testpeer.TCPBulkPayload()
	written, err := io.CopyN(app, bytes.NewReader(payload), int64(len(payload)))
	ul, dl, dropped, statsErr := ebpfRegistry.Stats()
	t.Logf("real production TCP write bytes=%d err=%v kernel UL=%d DL=%d ingressDrops=%d statsErr=%v", written, err, ul, dl, dropped, statsErr)
	require.NoError(t, err)
	reply := make([]byte, len(payload))
	_, err = io.ReadFull(app, reply)
	require.NoError(t, err)
	require.Equal(t, payload, reply)
	_, err = app.Write([]byte{1})
	require.NoError(t, err)
	ul, dl, _, err = ebpfRegistry.Stats()
	require.NoError(t, err)
	require.Greater(t, ul, uint64(2))
	require.Greater(t, dl, uint64(2))
	liveDevice, err := netlink.LinkByIndex(device.Attrs().Index)
	require.NoError(t, err)
	require.Equal(t, uint32(1), liveDevice.Attrs().GSOMaxSegs)
	t.Logf("real production TCP bulk bytes=%d uplink=%d downlink=%d endpointGSO=%d", len(payload), ul, dl, liveDevice.Attrs().GSOMaxSegs)
	require.NoError(t, app.Close())
	join()
	pdu.ReleaseTunnel()
	require.Nil(t, pdu.GetTunInterface())
	_, err = netlink.LinkByName(device.Attrs().Name)
	require.Error(t, err)
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	require.NoError(t, err)
	require.Empty(t, routes)
	rules, err := netlink.RuleList(netlink.FAMILY_V4)
	require.NoError(t, err)
	for _, rule := range rules {
		require.NotEqual(t, table, rule.Table)
	}
	// Released management port is immediately reusable; no worker can retain it.
	port, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("10.88.0.1"), Port: 2152})
	require.NoError(t, err)
	require.NoError(t, port.Close())
}
