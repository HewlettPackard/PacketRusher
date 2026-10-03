// SPDX-License-Identifier: Apache-2.0
//go:build linux

package ngap

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/ishidawataru/sctp"
	"golang.org/x/sys/unix"
)

// dialSCTPContext owns the descriptor throughout the handshake. The native
// SCTPConn API has no cancellable dial: abandoning its blocking dial in a
// goroutine leaves the socket bound until the kernel's INIT retries finish.
// Config endpoints each contain one address; multiple AMFs remain independent
// associations. Both IPv4 and IPv6 (including scoped IPv6) are supported.
func dialSCTPContext(ctx context.Context, local, remote netip.AddrPort) (_ *sctp.SCTPConn, err error) {
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if !local.IsValid() || !remote.IsValid() || remote.Port() == 0 {
		return nil, fmt.Errorf("invalid SCTP endpoints: local=%s remote=%s", local, remote)
	}
	family := unix.AF_INET
	if !local.Addr().Is4() || !remote.Addr().Is4() {
		family = unix.AF_INET6
	}
	loc, err := sctpSockaddr(local, family)
	if err != nil {
		return nil, err
	}
	rem, err := sctpSockaddr(remote, family)
	if err != nil {
		return nil, err
	}
	fd, err := unix.Socket(family, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_SCTP)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = unix.Close(fd)
		}
	}()
	// The pinned SCTPConn.Close only closes descriptors greater than zero.
	// A library user may legitimately have closed stdin: do not transfer fd 0
	// to that wrapper or shutdown would leak the socket and block reader joins.
	if fd == 0 {
		owned, dupErr := unix.FcntlInt(uintptr(fd), unix.F_DUPFD_CLOEXEC, 3)
		if dupErr != nil {
			return nil, dupErr
		}
		_ = unix.Close(fd)
		fd = owned
	}
	if family == unix.AF_INET6 {
		if err = unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 0); err != nil {
			return nil, err
		}
	}
	if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_BROADCAST, 1); err != nil {
		return nil, err
	}
	// Native struct sctp_initmsg is four uint16 values. Using a copied option
	// buffer avoids a nested uintptr pointer into a movable Go stack.
	options := make([]byte, 8)
	binary.NativeEndian.PutUint16(options[0:2], 2)
	binary.NativeEndian.PutUint16(options[2:4], 2)
	if err = unix.SetsockoptString(fd, unix.IPPROTO_SCTP, sctp.SCTP_INITMSG, string(options)); err != nil {
		return nil, err
	}
	if err = unix.Bind(fd, loc); err != nil {
		return nil, err
	}
	err = unix.Connect(fd, rem)
	if errors.Is(err, unix.EINPROGRESS) {
		err = awaitSCTPConnect(ctx, fd)
	}
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	// Existing receive/write paths use a blocking native SCTPConn. Ownership
	// transfers only after the handshake and cancellation check have succeeded.
	if err = unix.SetNonblock(fd, false); err != nil {
		return nil, err
	}
	return sctp.NewSCTPConn(fd, nil), nil
}

func awaitSCTPConnect(ctx context.Context, fd int) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		wait := 20 * time.Millisecond
		if deadline, ok := ctx.Deadline(); ok {
			wait = min(wait, time.Until(deadline))
		}
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLOUT}}
		_, err := unix.Poll(poll, max(1, int(wait.Milliseconds())))
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if poll[0].Revents == 0 {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		errno, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			return err
		}
		if errno != 0 {
			return unix.Errno(errno)
		}
		if _, err := unix.Getpeername(fd); err != nil {
			return err
		}
		return nil
	}
}

func sctpSockaddr(endpoint netip.AddrPort, family int) (unix.Sockaddr, error) {
	if family == unix.AF_INET {
		return &unix.SockaddrInet4{Port: int(endpoint.Port()), Addr: endpoint.Addr().As4()}, nil
	}
	address := &unix.SockaddrInet6{Port: int(endpoint.Port()), Addr: endpoint.Addr().As16()}
	if zone := endpoint.Addr().Zone(); zone != "" {
		iface, err := net.InterfaceByName(zone)
		if err != nil {
			return nil, fmt.Errorf("SCTP address zone %q: %w", zone, err)
		}
		address.ZoneId = uint32(iface.Index)
	}
	return address, nil
}
