/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/ngap"
	"net/netip"

	log "github.com/sirupsen/logrus"

	"github.com/ishidawataru/sctp"
)

const bufsize = 65535

func RunServer(ServerIpPort netip.AddrPort, fgc *context.Aio5gc) {
	ln, err := Listen(ServerIpPort)
	if err != nil {
		log.Errorf("[5GC] %v", err)
		return
	}
	fgc.RegisterCloser(ln.Close)
	Serve(ln, fgc)
}

// Serve accepts associations on l until it is closed.
func Serve(l *Listener, fgc *context.Aio5gc) {
	if !fgc.BeginWork() {
		return
	}
	defer fgc.EndWork()
	for {
		conn, err := l.Accept()
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

		// Each association has its own immutable connection owner. Replacing the
		// AMF's lookup entry cannot redirect replies from an old reader to a new socket.
		gnb := &context.GNBContext{}
		gnb.SetSCTPConn(conn.(*sctp.SCTPConn))
		_ = fgc.GetAMFContext().AddGnb(remote.String(), gnb)
		retire := fgc.Own(conn.Close)
		if !fgc.BeginWork() {
			_ = retire()
			fgc.GetAMFContext().RemoveGnb(remote.String(), gnb)
			continue
		}
		go func() {
			defer fgc.EndWork()
			defer fgc.GetAMFContext().RemoveGnb(remote.String(), gnb)
			defer retire()
			_ = listenAndServe(conn.(*sctp.SCTPConn), bufsize, gnb, fgc)
		}()

	}
}

func listenAndServe(conn *sctp.SCTPConn, bufsize int, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn
	for {
		n, err := conn.Read(buf)
		if err != nil {
			// The association is gone; the gNB may dial a new one, which Accept serves.
			log.Printf("[5GC] Read failed: %v", err)
			return err
		}
		if n > 0 {
			if err := ngap.Dispatch(buf[:n], gnb, fgc); err != nil {
				if fgc.Context().Err() == nil {
					fgc.RecordError(err)
					log.Error("[5GC] Dispatch failed: ", err)
				}
			}
		}
	}
}
