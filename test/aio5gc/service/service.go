/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/ngap"
	"net/netip"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/ishidawataru/sctp"
)

var bufsize = 65535

func RunServer(ServerIpPort netip.AddrPort, fgc *context.Aio5gc) {
	ln, err := Listen(ServerIpPort)
	if err != nil {
		log.Fatalf("[5GC] %v", err)
	}
	Serve(ln, fgc)
}

// Listener is the AMF's SCTP endpoint. A test that needs the AMF to refuse associations
// for a while closes it and later listens again.
type Listener struct {
	ln     *sctp.SCTPListener
	addr   *sctp.SCTPAddr
	closed atomic.Bool
}

func Listen(ServerIpPort netip.AddrPort) (*Listener, error) {
	addr, err := sctp.ResolveSCTPAddr("sctp", ServerIpPort.String())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve MockedAMF SCTP address %w", err)
	}
	ln, err := sctp.ListenSCTP("sctp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("[5GC] Listen on ", ln.Addr())
	return &Listener{ln: ln, addr: addr}, nil
}

// Close stops accepting associations. Closing an SCTP listener does not wake an accept
// blocked on it, and the socket keeps accepting until that accept returns, so one last
// association is dialled to release it; Serve closes it and returns.
func (l *Listener) Close() error {
	l.closed.Store(true)
	err := l.ln.Close()
	if wake, dialErr := sctp.DialSCTP("sctp", nil, l.addr); dialErr == nil {
		_ = wake.Close()
	}
	return err
}

// Serve accepts associations on l until it is closed.
func Serve(l *Listener, fgc *context.Aio5gc) {
	for {
		conn, err := l.ln.Accept()
		if err == nil && l.closed.Load() {
			_ = conn.Close()
		}
		if err != nil || l.closed.Load() {
			log.Info("[5GC] Stopped accepting: ", err)
			return
		}
		remote := conn.RemoteAddr()
		if remote == nil {
			// The gNB closed the association before it could be served.
			log.Info("[5GC] Accepted Connection closed by its peer")
			_ = conn.Close()
			continue
		}
		log.Info("[5GC] Accepted Connection from RemoteAddr: ", remote)

		amf := fgc.GetAMFContext()
		gnb, err := amf.GetGnb(remote.String())
		if err != nil {
			gnb = &context.GNBContext{}
			amf.AddGnb(remote.String(), gnb)
		}
		gnb.SetSCTPConn(conn.(*sctp.SCTPConn))
		go listenAndServe(conn.(*sctp.SCTPConn), bufsize, gnb, fgc)
	}
}

func listenAndServe(conn *sctp.SCTPConn, bufsize int, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn
	for {
		_, err := conn.Read(buf)
		if err != nil {
			// The association is gone; the gNB may dial a new one, which Accept serves.
			log.Printf("[5GC] Read failed: %v", err)
			return err
		}
		ngap.Dispatch(buf, gnb, fgc)
	}
}
