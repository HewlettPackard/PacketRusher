// SPDX-License-Identifier: Apache-2.0
//go:build linux

package service

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ishidawataru/sctp"
	"golang.org/x/sys/unix"
)

// Listener owns a nonblocking SCTP listening descriptor registered with Go's
// poller. File.Close wakes a blocked Accept without a synthetic association or
// a second dial, including when no gNB ever connected. Accepted associations
// remain native blocking SCTPConn instances used by the existing wire codecs.
type Listener struct {
	file       *os.File
	raw        syscall.RawConn
	addr       netip.AddrPort
	closed     atomic.Bool
	once       sync.Once
	closeErr   error
	accepting  chan struct{}
	acceptOnce sync.Once
}

func Listen(endpoint netip.AddrPort) (*Listener, error) {
	if !endpoint.IsValid() {
		return nil, fmt.Errorf("invalid AMF endpoint")
	}
	family := unix.AF_INET
	var address unix.Sockaddr
	if endpoint.Addr().Is4() {
		address = &unix.SockaddrInet4{Port: int(endpoint.Port()), Addr: endpoint.Addr().As4()}
	} else {
		family = unix.AF_INET6
		address = &unix.SockaddrInet6{Port: int(endpoint.Port()), Addr: endpoint.Addr().As16()}
	}
	fd, err := unix.Socket(family, unix.SOCK_STREAM|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, unix.IPPROTO_SCTP)
	if err != nil {
		return nil, fmt.Errorf("create AMF SCTP listener: %w", err)
	}
	// SCTPConn.SetInitMsg exposes the same native options as ListenSCTP; this
	// temporary wrapper never owns/ closes the listening descriptor.
	if err = sctp.NewSCTPConn(fd, nil).SetInitMsg(sctp.SCTP_MAX_STREAM, 0, 0, 0); err == nil {
		err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	}
	if err == nil {
		err = unix.Bind(fd, address)
	}
	if err == nil {
		err = unix.Listen(fd, unix.SOMAXCONN)
	}
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("bind AMF SCTP endpoint %s: %w", endpoint, err)
	}
	bound, err := unix.Getsockname(fd)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	switch bound := bound.(type) {
	case *unix.SockaddrInet4:
		endpoint = netip.AddrPortFrom(netip.AddrFrom4(bound.Addr), uint16(bound.Port))
	case *unix.SockaddrInet6:
		endpoint = netip.AddrPortFrom(netip.AddrFrom16(bound.Addr), uint16(bound.Port))
	}
	file := os.NewFile(uintptr(fd), "aio5gc-sctp-listener")
	raw, err := file.SyscallConn()
	if err != nil {
		file.Close()
		return nil, err
	}
	return &Listener{file: file, raw: raw, addr: endpoint, accepting: make(chan struct{})}, nil
}
func (l *Listener) Addr() netip.AddrPort { return l.addr }
func (l *Listener) Accept() (net.Conn, error) {
	var accepted int
	var acceptErr error
	err := l.raw.Read(func(fd uintptr) bool {
		for {
			accepted, _, acceptErr = unix.Accept4(int(fd), unix.SOCK_CLOEXEC)
			if errors.Is(acceptErr, unix.EINTR) {
				continue
			}
			waiting := errors.Is(acceptErr, unix.EAGAIN) || errors.Is(acceptErr, unix.EWOULDBLOCK)
			if waiting {
				l.acceptOnce.Do(func() { close(l.accepting) })
			}
			return !waiting
		}
	})
	if err != nil {
		return nil, err
	}
	if acceptErr != nil {
		return nil, acceptErr
	}
	return sctp.NewSCTPConn(accepted, nil), nil
}
func (l *Listener) Close() error {
	l.once.Do(func() { l.closed.Store(true); l.closeErr = l.file.Close() })
	return l.closeErr
}

// Accepting closes when Accept has reached the poller with an empty queue. It
// lets transport teardown tests exercise a blocked accept deterministically.
func (l *Listener) Accepting() <-chan struct{} { return l.accepting }
