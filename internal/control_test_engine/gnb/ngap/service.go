/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

// ConnCount offsets the local port of each association; it is advanced by every dial,
// including re-establishments, which may run concurrently for different AMFs.
var ConnCount atomic.Int32

const (
	reassociateInitialBackoff = time.Second
	reassociateMaxBackoff     = 30 * time.Second
	reassociateSetupTimeout   = 5 * time.Second
)

func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {
	if err := dialAmf(amf, gnb); err != nil {
		return err
	}

	log.Info("[GNB][SCTP] Starting GnbListen goroutine...")
	go GnbListen(amf, gnb)

	return nil
}

func dialAmf(amf *context.GNBAmf, gnb *context.GNBContext) error {

	// check AMF IP and AMF port.
	remote := amf.GetAmfIpPort().String()
	gnbAddrPort := gnb.GetGnbIpPort()
	local := netip.AddrPortFrom(gnbAddrPort.Addr(), gnbAddrPort.Port()+uint16(ConnCount.Add(1)-1)).String()

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

	return nil
}

// GnbListen reads the association with one AMF for as long as the gNB runs. When an
// association that had completed NG Setup is lost, the UEs served through it are released and
// the association is re-dialled with backoff and set up again, so an AMF restart costs the UEs
// it served rather than the whole gNB. An association that never completed NG Setup is left
// to InitGnb's own retries, as before.
func GnbListen(amf *context.GNBAmf, gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	var lostAt time.Time
	attempt := 0
	backoff := reassociateInitialBackoff

	for {
		conn := amf.GetSCTPConn()
		err := readAssociation(amf, gnb, conn, buf)
		if gnb.IsTerminated() {
			return
		}
		if !gnb.HasGnbAmf(amf.GetAmfId()) {
			// Abandoned by InitGnb's startup retries, or removed by an AMF
			// Configuration Update: not an association to re-establish, even if a late
			// NG Setup Response marked it active.
			return
		}

		if amf.GetState() == context.Active {
			// An established association was lost: start a new outage.
			amf.SetStateInactive()
			_ = conn.Close()
			released := gnb.ReleaseUesOfAmf(amf.GetAmfId())
			lostAt = time.Now()
			attempt = 0
			backoff = reassociateInitialBackoff
			log.Error("[GNB][SCTP] Association with AMF ", amf.GetAmfIpPort(), " lost (", err, "); released ",
				released, " UE contexts served through it, re-establishing")
		} else if lostAt.IsZero() {
			// Never set up: not ours to re-establish.
			log.Warn("[GNB][SCTP] Association with AMF ", amf.GetAmfIpPort(), " closed before NG Setup completed: ", err)
			return
		} else {
			// A re-establishment attempt that did not complete NG Setup.
			_ = conn.Close()
			time.Sleep(backoff)
			backoff = min(2*backoff, reassociateMaxBackoff)
		}

		// Dial until an association exists, then set it up; the setup response arrives
		// through readAssociation on the next pass.
		for {
			// Checked before every dial: Terminate closes only the association it can see,
			// so one dialled after it would never be closed.
			if gnb.IsTerminated() {
				return
			}
			attempt++
			if err := dialAmf(amf, gnb); err == nil {
				break
			}
			log.Error("[GNB][SCTP] Still no association with AMF ", amf.GetAmfIpPort(), " after ",
				time.Since(lostAt).Round(time.Second), " (attempt ", attempt, "); retrying in ", backoff)
			time.Sleep(backoff)
			backoff = min(2*backoff, reassociateMaxBackoff)
		}
		trigger.SendNgSetupRequest(gnb, amf)
		go awaitReassociation(amf, amf.GetSCTPConn(), lostAt, attempt)
	}
}

// readAssociation dispatches NGAP messages from conn until reading it fails.
func readAssociation(amf *context.GNBAmf, gnb *context.GNBContext, conn *sctp.SCTPConn, buf []byte) error {
	for {
		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			return err
		}

		log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go Dispatch(amf, gnb, forwardData)
	}
}

// awaitReassociation reports a re-established association once NG Setup completes, and closes
// it if setup does not complete in time so that GnbListen tries again.
func awaitReassociation(amf *context.GNBAmf, conn *sctp.SCTPConn, lostAt time.Time, attempt int) {
	deadline := time.After(reassociateSetupTimeout)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if amf.GetState() == context.Active {
				log.Warn("[GNB][SCTP] Association with AMF ", amf.GetAmfIpPort(), " re-established after ",
					time.Since(lostAt).Round(time.Second), " (attempt ", attempt, ")")
				return
			}
		case <-deadline:
			if amf.CloseIfNotActive(conn) {
				log.Error("[GNB][SCTP] NG Setup with AMF ", amf.GetAmfIpPort(), " did not complete within ",
					reassociateSetupTimeout, " (attempt ", attempt, "); retrying")
			}
			return
		}
	}
}
