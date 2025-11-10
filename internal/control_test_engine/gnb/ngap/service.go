/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"time"

	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

var ConnCount int

func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {

	// check AMF IP and AMF port.
	remote := amf.GetAmfIpPort().String()
	gnbAddrPort := gnb.GetGnbIpPort()
	local := netip.AddrPortFrom(gnbAddrPort.Addr(), gnbAddrPort.Port()+uint16(ConnCount)).String()
	ConnCount++

	log.Info("[GNB][SCTP] Initializing connection: local=", local, " remote=", remote)

	rem, err := sctp.ResolveSCTPAddr("sctp", remote)
	if err != nil {
		log.Error("[GNB][SCTP] Failed to resolve remote address: ", err)
		return err
	}
	loc, err := sctp.ResolveSCTPAddr("sctp", local)
	if err != nil {
		log.Error("[GNB][SCTP] Failed to resolve local address: ", err)
		return err
	}

	log.Info("[GNB][SCTP] Attempting SCTP dial...")

	// streams := amf.GetTNLAStreams()

	// Wrap SCTP dial in a goroutine with timeout
	type dialResult struct {
		conn *sctp.SCTPConn
		err  error
	}
	dialChan := make(chan dialResult, 1)

	go func() {
		conn, err := sctp.DialSCTPExt(
			"sctp",
			loc,
			rem,
			sctp.InitMsg{NumOstreams: 2, MaxInstreams: 2})
		dialChan <- dialResult{conn: conn, err: err}
	}()

	// Wait for dial with 5-second timeout
	var conn *sctp.SCTPConn
	select {
	case result := <-dialChan:
		conn = result.conn
		err = result.err
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("SCTP dial timeout after 5 seconds")
		log.Error("[GNB][SCTP] SCTP dial timeout")
		return err
	}

	if err != nil {
		log.Error("[GNB][SCTP] SCTP dial failed: ", err)
		amf.SetSCTPConn(nil)
		return err
	}

	log.Info("[GNB][SCTP] SCTP connection established successfully")

	// set streams and other information about TNLA

	// successful established SCTP (TNLA - N2)
	amf.SetSCTPConn(conn)
	gnb.SetN2(conn)

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

	log.Info("[GNB][SCTP] Starting GnbListen goroutine...")
	go GnbListen(amf, gnb)

	return nil
}

func GnbListen(amf *context.GNBAmf, gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := amf.GetSCTPConn()

	/*
		defer func() {
			err := conn.Close()
			if err != nil {
				log.Info("[GNB][SCTP] Error in closing SCTP association for %d AMF\n", amf.GetAmfId())
			}
		}()
	*/

	for {

		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			break
		}

		log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go Dispatch(amf, gnb, forwardData)

	}

}
