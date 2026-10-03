/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	stdcontext "context"
	"errors"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"net/netip"
	"sync/atomic"
	"time"

	ngapmsg "github.com/free5gc/ngap/message"
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

// ConnCount offsets the local port of each AMF's association. It is advanced once per
// AMF, at its first dial; a re-establishment reuses that AMF's port. Advancing it on
// every redial walked the port into the range of gNB processes configured on the ports
// next to this one, and each collision cost a failed dial and a longer backoff.
var ConnCount atomic.Int32

// DialCount counts every dial, including re-establishments.
var DialCount atomic.Int32

const (
	reassociateInitialBackoff = time.Second
	reassociateMaxBackoff     = 30 * time.Second
	reassociateSetupTimeout   = 5 * time.Second
	processingQueueTimeout    = 5 * time.Second
)

// errAssociationUnwanted is returned by dialAmf when the gNB was terminated, or stopped
// serving through the AMF, while the dial was in progress.
var errAssociationUnwanted = errors.New("association no longer wanted")

// InitConn preserves the original API with a bounded dial. Context-aware
// constructors use InitConnContext to share their total startup budget.
func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {
	return InitConnContext(stdcontext.Background(), amf, gnb)
}

func InitConnContext(ctx stdcontext.Context, amf *context.GNBAmf, gnb *context.GNBContext) error {
	if err := dialAmfContext(ctx, amf, gnb); err != nil {
		return err
	}
	if !gnb.RunAssociation(func() { GnbListen(amf, gnb) }) {
		_ = amf.GetSCTPConn().Close()
		return errAssociationUnwanted
	}
	return nil
}

func dialAmf(amf *context.GNBAmf, gnb *context.GNBContext) error {
	return dialAmfContext(stdcontext.Background(), amf, gnb)
}

func dialAmfContext(parent stdcontext.Context, amf *context.GNBAmf, gnb *context.GNBContext) error {
	ctx, cancel := stdcontext.WithTimeout(parent, 5*time.Second)
	// The small cancellation watcher is joined before returning; the actual
	// native dial has no background socket/goroutine left after cancellation.
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-gnb.Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	defer func() { cancel(); <-watcherDone }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if gnb.IsTerminated() || !gnb.HasGnbAmf(amf.GetAmfId()) {
		return errAssociationUnwanted
	}
	remote := amf.GetAmfIpPort()
	gnbAddrPort := gnb.GetGnbIpPort()
	port := amf.GetLocalPort()
	if port == 0 {
		port = localAssociationPort(gnbAddrPort.Port(), ConnCount.Add(1)-1)
		amf.SetLocalPort(port)
	}
	DialCount.Add(1)
	local := netip.AddrPortFrom(gnbAddrPort.Addr(), port)
	log.Info("[GNB][SCTP] Initializing connection: local=", local, " remote=", remote)
	conn, err := dialSCTPContext(ctx, local, remote)
	if err != nil {
		return fmt.Errorf("SCTP dial local=%s remote=%s: %w", local, remote, err)
	}
	// Install receive metadata before publishing or starting a native reader.
	if err := conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO); err != nil {
		_ = conn.Close()
		return fmt.Errorf("subscribe SCTP data events: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return err
	}
	if gnbAddrPort.Port() == 0 {
		if addr, ok := conn.LocalAddr().(*sctp.SCTPAddr); ok {
			amf.SetLocalPort(uint16(addr.Port))
		}
	}
	if !gnb.PublishAssociation(amf, conn) {
		_ = conn.Close()
		return errAssociationUnwanted
	}
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

		// Retire this reader's transport before retry/backoff, preserving a
		// replacement published meanwhile.
		amf.ClearSCTPConn(conn)

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
			amf.RejectSetup(fmt.Errorf("SCTP association closed before NG Setup: %w", err))
			return
		} else {
			// A re-establishment attempt that did not complete NG Setup.
			_ = conn.Close()
			if !waitBackoff(gnb.Done(), backoff) {
				return
			}
			backoff = min(2*backoff, reassociateMaxBackoff)
		}

		// Dial until an association exists, then set it up; the setup response arrives
		// through readAssociation on the next pass.
		for {
			// Checked before every dial so as not to dial for nothing; dialAmf checks
			// again as it publishes, since either can change while it dials.
			if gnb.IsTerminated() || !gnb.HasGnbAmf(amf.GetAmfId()) {
				return
			}
			attempt++
			err := dialAmf(amf, gnb)
			if err == nil {
				break
			}
			if errors.Is(err, errAssociationUnwanted) {
				return
			}
			log.Error("[GNB][SCTP] Still no association with AMF ", amf.GetAmfIpPort(), " after ",
				time.Since(lostAt).Round(time.Second), " (attempt ", attempt, "); retrying in ", backoff)
			if !waitBackoff(gnb.Done(), backoff) {
				return
			}
			backoff = min(2*backoff, reassociateMaxBackoff)
		}
		trigger.SendNgSetupRequest(gnb, amf)
		reassociatedConn, outage, completedAttempt := amf.GetSCTPConn(), lostAt, attempt
		gnb.RunAssociation(func() { awaitReassociationContext(gnb.Done(), amf, reassociatedConn, outage, completedAttempt) })
	}
}

// readAssociation dispatches NGAP messages from conn until reading it fails.
func readAssociation(amf *context.GNBAmf, gnb *context.GNBContext, conn *sctp.SCTPConn, buf []byte) error {
	dispatcher := newOrderedDispatcher(4096)
	defer func() {
		_ = conn.Close() // Also interrupt uplink writes holding a UE processing lock.
		// A lost association releases all of its UEs. Cancel blocked delivery before
		// waiting for workers; otherwise a full UE channel could prevent reassociation.
		gnb.GetUePool().Range(func(_, value any) bool {
			ue := value.(*context.GNBUe)
			if ue.GetAmfId() == amf.GetAmfId() {
				ue.FailUEChannel()
			}
			return true
		})
		dispatcher.stop()
		dispatcher.wait()
	}()
	for {
		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			return err
		}

		log.Info("[GNB][SCTP] Receive message in ", info.Stream, " stream\n")

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// Decode before enqueueing: launching a goroutine per packet can reorder
		// a NAS PDU and the following UE Context Release Command.
		message, err := ngapmsg.Parse(forwardData)
		if err != nil || message == nil {
			log.Error("[GNB][NGAP] Cannot decode received message: ", err)
			continue
		}
		key := dispatcher.messageKey(gnb, message)
		queueContext, cancelQueue := stdcontext.WithTimeout(stdcontext.Background(), processingQueueTimeout)
		accepted := dispatcher.enqueueContext(queueContext, key, func() {
			dispatchUEMessage(amf, gnb, message)
		})
		cancelQueue()
		if !accepted {
			return fmt.Errorf("NGAP processing queue stalled for %s", processingQueueTimeout)
		}
	}
}

// awaitReassociation reports a re-established association once NG Setup completes, and closes
// it if setup does not complete in time so that GnbListen tries again.
func awaitReassociation(amf *context.GNBAmf, conn *sctp.SCTPConn, lostAt time.Time, attempt int) {
	awaitReassociationContext(nil, amf, conn, lostAt, attempt)
}

func awaitReassociationContext(stopped <-chan struct{}, amf *context.GNBAmf, conn *sctp.SCTPConn, lostAt time.Time, attempt int) {
	deadline := time.After(reassociateSetupTimeout)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			// A later attempt has replaced this association: the outcome is that
			// attempt's to report, not this one's.
			if amf.GetSCTPConn() != conn {
				return
			}
			if amf.GetState() == context.Active {
				log.Warn("[GNB][SCTP] Association with AMF ", amf.GetAmfIpPort(), " re-established after ",
					time.Since(lostAt).Round(time.Second), " (attempt ", attempt, ")")
				return
			}
		case <-stopped:
			return
		case <-deadline:
			if amf.CloseIfNotActive(conn) {
				log.Error("[GNB][SCTP] NG Setup with AMF ", amf.GetAmfIpPort(), " did not complete within ",
					reassociateSetupTimeout, " (attempt ", attempt, "); retrying")
			}
			return
		}
	}
}

func waitBackoff(stopped <-chan struct{}, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-stopped:
		return false
	case <-timer.C:
		return true
	}
}

// Configured port zero requests a kernel-assigned endpoint for every association;
// counting associations must not turn it into a privileged fixed port.
func localAssociationPort(configured uint16, offset int32) uint16 {
	if configured == 0 {
		return 0
	}
	return configured + uint16(offset)
}
