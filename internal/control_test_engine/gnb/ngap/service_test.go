/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 * © Copyright 2026 Valentin D'Emmanuele
 */
package ngap

import (
	log "my5G-RANTester/internal/log"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func reestablishedLogs(logs *observer.ObservedLogs) int {
	n := 0
	for _, e := range logs.All() {
		if e.Level == zap.WarnLevel && strings.Contains(e.Message, "re-established") {
			n++
		}
	}
	return n
}

// An attempt whose association died before NG Setup is followed by another. When the
// later one sets the AMF up, only it may report the association as re-established; the
// earlier watcher used to see the AMF Active and report it too.
func TestAwaitReassociationLeavesALaterAttemptToReport(t *testing.T) {
	core, hook := observer.New(zap.WarnLevel)
	t.Cleanup(log.Replace(core))

	gnb := createTestGNBContext()
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38499"))
	first, second := &sctp.SCTPConn{}, &sctp.SCTPConn{}
	amf.SetSCTPConn(second)
	amf.SetStateActive()

	done := make(chan struct{})
	go func() { awaitReassociation(amf, first, time.Now(), 1); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("the superseded watcher should stop")
	}
	assert.Zero(t, reestablishedLogs(hook), "a superseded attempt must not report the re-establishment")

	awaitReassociation(amf, second, time.Now(), 2)
	assert.Equal(t, 1, reestablishedLogs(hook), "the attempt that set the AMF up should report it")
}
