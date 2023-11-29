/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */

package amf

import (
	"fmt"
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapSctp"
	"my5G-RANTester/test/amf/context"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/ishidawataru/sctp"
)

var bufsize = 65535

func (a *Amf) runServer(ServerIp string, ServerPort int) {
	addr, err := sctp.ResolveSCTPAddr("sctp", fmt.Sprintf("%s:%d", ServerIp, ServerPort))
	if err != nil {
		log.Fatalf("[AMF] Failed to resolve MockedAMF SCTP address %v", err)
	}
	ln, err := sctp.ListenSCTP("sctp", addr)
	if err != nil {
		log.Fatalf("[AMF] Failed to listen: %v", err)
	}
	log.Info("[AMF] Listen on ", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("[AMF] failed to accept: %v", err)
		}
		log.Info("[AMF] Accepted Connection from RemoteAddr: ", conn.RemoteAddr())

		gnb, err := a.GetGnb(conn.RemoteAddr().String())
		if err != nil {
			gnb = context.GNBContext{}
			a.context.AddGnb(conn.RemoteAddr().String(), gnb)
		}

		go a.serveClient(conn.(*sctp.SCTPConn), bufsize, gnb)
	}
}

func (a Amf) serveClient(conn *sctp.SCTPConn, bufsize int, gnb context.GNBContext) error {
	buf := make([]byte, bufsize+128) // add overhead of SCTPSndRcvInfoWrappedConn
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("[AMF] Read failed: %v", err)
			os.Exit(1)
		}
		ngapReq, err := ngap.Decoder(buf)
		if err != nil {
			log.Printf("[AMF] Ngap Decode failed: %v", err)
			os.Exit(1)
		}
		res, err := a.ngapReqHandler(ngapReq, gnb)
		if err != nil {
			log.Printf("[AMF] Req handling failed: %v", err)
			os.Exit(1)
		}
		info := &sctp.SndRcvInfo{
			Stream: uint16(0),
			PPID:   ngapSctp.NGAP_PPID,
		}
		_, err = conn.SCTPWrite(res, info)
		if err != nil {
			log.Printf("[AMF] write failed: %v", err)
			os.Exit(1)
		}
	}
}
