// SPDX-License-Identifier: Apache-2.0
package ue

import (
	"fmt"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	ngapTrigger "my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	nasTrigger "my5G-RANTester/internal/control_test_engine/ue/nas/trigger"
)

func attachment(r *procedures.ControlRequest, ue *context.UEContext) procedures.Attachment {
	a := procedures.Attachment{UE: int(ue.GetPrUeId()), Generation: r.Generation,
		ConnectionGeneration: ue.ConnectionGeneration(), Connected: ue.GetGnbTx() != nil,
		ActivePDUSessions: []uint8{}}
	a.State = map[int]string{context.MM5G_NULL: "starting", context.MM5G_DEREGISTERED: "deregistered",
		context.MM5G_REGISTERED_INITIATED: "registering", context.MM5G_REGISTERED: "registered",
		context.MM5G_IDLE: "idle", context.MM5G_SERVICE_REQ_INIT: "reconnecting",
		context.MM5G_DEREGISTERED_INIT: "deregistering"}[ue.GetStateMM()]
	for id, node := range r.Gnbs {
		if node.GetInboundChannel() == ue.GetGnbInboundChannel() {
			a.GNB = id
			if gu, err := node.GetGnbUeByPrUeId(ue.GetPrUeId()); err == nil {
				a.Ready = gu.GetState() == gnb.Ready && node.NGSetupReady()
			}
			break
		}
	}
	expectedActive := 0
	for id := uint8(1); id <= 15; id++ {
		if pdu, err := ue.GetPduSession(id); err == nil && pdu.GetStateSM() == context.SM5G_PDU_SESSION_ACTIVE {
			a.ActivePDUSessions = append(a.ActivePDUSessions, id)
			if int(id) <= r.ExpectedPDUSessions {
				expectedActive++
			}
		}
	}
	a.Ready = a.Ready && a.Connected && a.State == "registered" && !ue.ControlHandoverTarget().IsValid() && expectedActive == r.ExpectedPDUSessions
	return a
}

func handleControl(r *procedures.ControlRequest, ue *context.UEContext) bool {
	if r == nil {
		return true
	}
	a := attachment(r, ue)
	if err := r.Context.Err(); err != nil {
		r.Respond(a, err)
		return true
	}
	if r.ExpectedGeneration != 0 && r.ExpectedGeneration != a.Generation ||
		r.ExpectedConnection != 0 && r.ExpectedConnection != a.ConnectionGeneration {
		r.Respond(a, procedures.ErrGeneration)
		return true
	}
	if r.Action == "inspect" {
		r.Respond(a, nil)
		return true
	}
	ready := a.Ready
	if r.Action == "deregister" {
		ready = a.State == "registered" && a.Connected
	}
	if r.Action == "reconnect" {
		ready = a.State == "idle" && !a.Connected
	}
	if !ready && r.Action != "terminate-after-timeout" {
		r.Respond(a, procedures.ErrNotReady)
		return true
	}
	var err error
	keep := true
	switch r.Action {
	case "terminate", "terminate-after-timeout":
		keep = ueMgrHandler(procedures.UeTesterMessage{Type: procedures.Terminate}, ue)
	case "deregister":
		// Switch-off deregistration has no NAS acknowledgement. Ending this UE
		// releases its local session resources; the scenario parks until register.
		nasTrigger.InitDeregistration(ue)
		r.DidDeregister = true
		r.LastGNB = a.GNB
		keep = false
	case "idle":
		ueMgrHandler(procedures.UeTesterMessage{Type: procedures.Idle}, ue)
	case "reconnect":
		ueMgrHandler(procedures.UeTesterMessage{Type: procedures.ServiceRequest}, ue)
	case "xn-handover", "ng-handover":
		source, target := r.Gnbs[a.GNB], r.Gnbs[r.Target]
		if target == nil || target == source {
			err = fmt.Errorf("handover target must be another configured gNB")
		} else if !target.NGSetupReady() {
			err = procedures.ErrNotReady
		} else if r.Action == "ng-handover" {
			err = ngapTrigger.StartNgapHandover(source, target, ue.GetPrUeId())
		} else {
			var message gnb.UEMessage
			message, err = ngapTrigger.PrepareXnHandover(r.Context, source, target, ue.GetPrUeId())
			if err == nil {
				gnbMsgHandler(message, ue)
			}
		}
		if err == nil {
			sourceUE, _ := source.GetGnbUeByPrUeId(ue.GetPrUeId())
			ue.BeginControlHandover(target.GetN3GnbIp(), r.Context.Done(), sourceUE)
		}
	default:
		err = fmt.Errorf("unknown control action %q", r.Action)
	}
	r.Respond(attachment(r, ue), err)
	return keep
}
