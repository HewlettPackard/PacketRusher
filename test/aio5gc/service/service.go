/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package service

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/ngap"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/ishidawataru/sctp"
)

var bufsize = 65535

func RunServer(ServerIp string, ServerPort int, fgc *context.Aio5gc) {
	addr, err := sctp.ResolveSCTPAddr("sctp", fmt.Sprintf("%s:%d", ServerIp, ServerPort))
	if err != nil {
		log.Fatalf("[5GC] Failed to resolve MockedAMF SCTP address %v", err)
	}
	ln, err := sctp.ListenSCTP("sctp", addr)
	if err != nil {
		log.Fatalf("[5GC] Failed to listen: %v", err)
	}
	log.Info("[5GC] Listen on ", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("[5GC] failed to accept: %v", err)
		}
		log.Info("[5GC] Accepted Connection from RemoteAddr: ", conn.RemoteAddr())

		amf := fgc.GetAMFContext()
		gnb, err := amf.GetGnb(conn.RemoteAddr().String())
		if err != nil {
			gnb = &context.GNBContext{}
			amf.AddGnb(conn.RemoteAddr().String(), gnb)
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
			log.Printf("[5GC] Read failed: %v", err)
			os.Exit(1)
		}
		ngap.Dispatch(buf, gnb, fgc)
	}
}
