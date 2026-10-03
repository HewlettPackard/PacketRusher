/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Forsway Scandinavia AB
 */
package ue

import (
	"github.com/free5gc/nas/ie"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"
)

// Production loop dispatch must consume typed retries as well as generic work.
func TestHandleUERunsSessionRetry(t *testing.T) {
	capability := &ie.UESecCapability{Length: 2, EA05G: true, IA05G: true}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	uplink := make(chan gnbContext.UEMessage, 1)
	ue.SetGnbRx(uplink)
	session, err := ue.CreatePDUSession()
	if err != nil {
		t.Fatal(err)
	}
	session.SetStateSM_PDU_SESSION_PENDING()
	if !ue.SchedulePduSessionRetry(session) {
		t.Fatal("pending session should schedule a retry")
	}

	mgr := make(chan procedures.UeTesterMessage)
	stopped := make(chan struct{})
	go func() { handleUE(ue, mgr); close(stopped) }()
	t.Cleanup(func() {
		close(mgr)
		select {
		case <-stopped:
			ue.Terminate()
		case <-time.After(2 * time.Second):
			t.Error("the UE's goroutine should stop")
		}
	})

	select {
	case message := <-uplink:
		if !message.IsNas || len(message.Nas) == 0 {
			t.Fatal("the UE loop should encode and send the retried NAS request")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the UE loop did not execute the scheduled retry")
	}
	ue.Lock()
	retries := session.T3580Retries
	ue.Unlock()
	if retries != 1 {
		t.Fatalf("executed retry count = %d, want 1", retries)
	}
}
