/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package procedures

import "my5G-RANTester/internal/control_test_engine/gnb/context"

type UeTesterMessageType int32

const (
	Registration      UeTesterMessageType = 0
	Deregistration    UeTesterMessageType = 1
	NewPDUSession     UeTesterMessageType = 2
	DestroyPDUSession UeTesterMessageType = 3
	Terminate         UeTesterMessageType = 4
	Kill              UeTesterMessageType = 5
	Idle              UeTesterMessageType = 6
	ServiceRequest    UeTesterMessageType = 7
	Control           UeTesterMessageType = 8
)

type UeTesterMessage struct {
	Type    UeTesterMessageType
	Param   uint8
	GnbChan chan context.UEMessage
	Control *ControlRequest
}

// ControlRequest is an action requested on the control socket for one UE. The UE
// reports its status to its scenario, which runs the action and answers on Reply.
type ControlRequest struct {
	Action string
	Target string
	Reply  chan ControlReply
}

type ControlReply struct {
	Status UeStatus
	Err    error
}

// UeStatus is what the control socket reports about a UE.
type UeStatus struct {
	UeId        int    `json:"ue"`
	State       string `json:"state"`
	GnbId       string `json:"gnb,omitempty"`
	Connected   bool   `json:"connected"`
	Ready       bool   `json:"ready"`
	PduSessions []int  `json:"active_pdu_sessions"`
	// The UE only knows the channel of its gNB. Its scenario names that gNB, and
	// completes Ready with the number of PDU sessions it requested.
	GnbInboundChannel chan context.UEMessage `json:"-"`
}
