//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/netip"
	"sync"
	"time"
)

// managementSocket owns a joined Echo/control worker. Kernel-completed traffic
// stays in TCX; checksum-uncertain, reassembled or optional-header datagrams reach this socket
// only after normal UDP checksum validation, then strict session admission.
type managementSocket struct {
	conn    *net.UDPConn
	done    chan struct{}
	writeMu sync.Mutex
}

func newManagementSocket(local netip.Addr, allowed func(netip.AddrPort) bool, handlers ...func(netip.AddrPort, []byte)) (*managementSocket, error) {
	conn, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.AddrPortFrom(local, 2152)))
	if err != nil {
		return nil, err
	}
	s := &managementSocket{conn: conn, done: make(chan struct{})}
	go func() {
		defer close(s.done)
		var packet [65535]byte
		for {
			n, peer, err := conn.ReadFromUDPAddrPort(packet[:])
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Warn("[UE][eBPF] GTP Echo worker stopped: ", err)
				}
				return
			}
			if !allowed(peer) {
				continue
			}
			if response := echoResponse(packet[:n]); response != nil {
				if err := s.Send(response, peer); err != nil {
					log.Warnf("[UE][eBPF] Echo response from %s to %s failed: %v", conn.LocalAddr(), peer, err)
				}
			} else if len(handlers) > 0 {
				handlers[0](peer, packet[:n])
			}
		}
	}()
	return s, nil
}
func (s *managementSocket) Send(packet []byte, peer netip.AddrPort) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := s.conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	_, err := s.conn.WriteToUDPAddrPort(packet, peer)
	return err
}
func (s *managementSocket) Close() error {
	err := s.conn.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	<-s.done
	return nil
}
func echoResponse(packet []byte) []byte {
	if len(packet) < 12 || packet[0] != 0x32 || packet[1] != 1 || int(binary.BigEndian.Uint16(packet[2:4])) != len(packet)-8 || binary.BigEndian.Uint32(packet[4:8]) != 0 || packet[10] != 0 || packet[11] != 0 {
		return nil
	}
	// Recovery IE: this process has not restarted within its registry lifetime.
	return []byte{0x32, 2, 0, 6, 0, 0, 0, 0, packet[8], packet[9], 0, 0, 14, 0}
}
