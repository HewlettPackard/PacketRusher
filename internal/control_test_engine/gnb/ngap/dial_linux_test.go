// SPDX-License-Identifier: Apache-2.0
//go:build linux

package ngap

import (
	"context"
	"net/netip"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestCancelledNativeDialNeverCreatesSocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	conn, err := dialSCTPContext(ctx, netip.AddrPort{}, netip.AddrPort{})
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, conn)
}

func TestNativeDialWithClosedStdinOwnsClosableDescriptor(t *testing.T) {
	if os.Getenv("PACKETRUSHER_TEST_CLOSED_STDIN") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeDialWithClosedStdinOwnsClosableDescriptor$", "-test.v", "-test.timeout=5s")
		cmd.Env = append(os.Environ(), "PACKETRUSHER_TEST_CLOSED_STDIN=1")
		output, err := cmd.CombinedOutput()
		t.Logf("owned child with closed stdin:\n%s", output)
		require.NoError(t, err)
		return
	}
	// Make the real listener before releasing descriptor zero. Only the owned
	// child closes stdin; the parent runner's descriptors remain untouched.
	listener, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	require.NoError(t, err)
	defer unix.Close(listener)
	require.NoError(t, unix.Bind(listener, &unix.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}))
	require.NoError(t, unix.Listen(listener, 1))
	bound, err := unix.Getsockname(listener)
	require.NoError(t, err)
	remote := netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(bound.(*unix.SockaddrInet4).Port))
	require.NoError(t, unix.Close(0))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := dialSCTPContext(ctx, netip.MustParseAddrPort("127.0.0.1:0"), remote)
	require.NoError(t, err)
	defer conn.Close()
	_, err = unix.FcntlInt(0, unix.F_GETFD, 0)
	require.ErrorIs(t, err, unix.EBADF, "the native wrapper must not retain descriptor zero")
	peer, _, err := unix.Accept4(listener, unix.SOCK_CLOEXEC)
	require.NoError(t, err)
	defer unix.Close(peer)
	require.NoError(t, conn.Close(), "successful startup must transfer a descriptor SCTPConn can close")
}

// Run only in a dedicated namespace with dummy 198.18.0.2/24 up and no peer.
// Unlike a closed loopback port, the pending INIT cannot finish immediately.
func TestNativePendingSCTPDialCancellationReleasesBoundPort(t *testing.T) {
	if os.Getenv("PACKETRUSHER_SCTP_DIAL_BLACKHOLE") != "1" {
		t.Skip("requires owned blackhole network namespace")
	}
	before, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	local := netip.MustParseAddrPort("198.18.0.2:39012")
	go func() {
		conn, err := dialSCTPContext(ctx, local, netip.MustParseAddrPort("198.18.0.1:38412"))
		if conn != nil {
			_ = conn.Close()
		}
		result <- err
	}()
	observe, stopObservation := context.WithTimeout(context.Background(), time.Second)
	defer stopObservation()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		state, err := os.ReadFile("/proc/net/sctp/assocs")
		require.NoError(t, err)
		if strings.Contains(string(state), "39012") {
			t.Logf("actual pending association before cancellation:\n%s", state)
			break
		}
		select {
		case err := <-result:
			t.Fatalf("native handshake did not stay pending: %v", err)
		case <-observe.Done():
			t.Fatal("pending native association was not observed")
		case <-tick.C:
		}
	}
	cancelledAt := time.Now()
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(250 * time.Millisecond):
		t.Fatal("native dial ignored cancellation")
	}
	t.Logf("actual pending dial canceled in %s", time.Since(cancelledAt))
	after, err := os.ReadDir("/proc/self/fd")
	require.NoError(t, err)
	require.Len(t, after, len(before), "cancellation must close the dial descriptor")
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	require.NoError(t, err)
	defer unix.Close(fd)
	address, err := sctpSockaddr(local, unix.AF_INET)
	require.NoError(t, err)
	require.NoError(t, unix.Bind(fd, address), "the exact canceled endpoint must be immediately reusable")
}

func TestSCTPAddressesPreserveIPv6AndMappedIPv4(t *testing.T) {
	for _, text := range []string{"[::1]:38412", "127.0.0.1:38412"} {
		endpoint := netip.MustParseAddrPort(text)
		address, err := sctpSockaddr(endpoint, unix.AF_INET6)
		require.NoError(t, err)
		v6 := address.(*unix.SockaddrInet6)
		require.Equal(t, int(endpoint.Port()), v6.Port)
		require.Equal(t, endpoint.Addr().As16(), v6.Addr)
	}
}
