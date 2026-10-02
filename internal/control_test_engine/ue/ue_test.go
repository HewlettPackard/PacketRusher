/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package ue

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/sidf"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"testing"
	"time"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
)

// Work handed to the UE with RunOnUE is run by the UE's own goroutine, between the
// messages it handles.
func TestHandleUERunsHandedOverWork(t *testing.T) {
	capability := nasType.NewUESecurityCapability(nasMessage.RegistrationRequestUESecurityCapabilityType)
	capability.SetLen(2)
	capability.Buffer = []uint8{0x80, 0x80}
	ue := &context.UEContext{}
	ue.NewRanUeContext("0000000001", capability, "", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)

	mgr := make(chan procedures.UeTesterMessage)
	stopped := make(chan struct{})
	go func() { handleUE(ue, mgr); close(stopped) }()

	ran := make(chan struct{})
	handed := make(chan bool, 1)
	go func() { handed <- ue.RunOnUE(func() { close(ran) }) }()
	select {
	case <-ran:
	case <-time.After(2 * time.Second):
		t.Fatal("the UE's goroutine should run work handed to it")
	}
	if !<-handed {
		t.Fatal("RunOnUE should report the work as handed over")
	}

	close(mgr)
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("the UE's goroutine should stop when the scenario closes its channel")
	}
}
