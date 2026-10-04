// SPDX-License-Identifier: Apache-2.0
package procedures

import (
	"context"
	"encoding/json"
	"errors"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
)

const Control UeTesterMessageType = 8

var (
	ErrNotReady   = errors.New("UE is not ready for this action")
	ErrGeneration = errors.New("UE connection or registration generation changed")
	ErrStopped    = errors.New("UE simulation has stopped")
)

// Attachment contains local identifiers, never subscriber keys or IMSIs. It is
// read on the UE event loop, alongside the NAS state transitions it describes.
type Attachment struct {
	UE                   int               `json:"ue"`
	Generation           uint64            `json:"generation"`
	ConnectionGeneration uint64            `json:"connection_generation"`
	State                string            `json:"state"`
	GNB                  string            `json:"gnb,omitempty"`
	Connected            bool              `json:"connected"`
	Ready                bool              `json:"ready"`
	ActivePDUSessions    []uint8           `json:"active_pdu_sessions"`
	Tunnels              []TunnelSelection `json:"tunnels,omitempty"`
}

type TunnelSelection struct {
	PDU      uint8  `json:"pdu"`
	Backend  string `json:"backend"`
	Fallback string `json:"fallback,omitempty"`
}

// MarshalJSON exposes session identities as numbers. encoding/json otherwise
// treats []uint8 as bytes and writes a base64 string, including for zero sessions.
func (a Attachment) MarshalJSON() ([]byte, error) {
	type attachmentJSON Attachment
	sessions := make([]int, len(a.ActivePDUSessions))
	for i, id := range a.ActivePDUSessions {
		sessions[i] = int(id)
	}
	return json.Marshal(struct {
		attachmentJSON
		ActivePDUSessions []int `json:"active_pdu_sessions"`
	}{attachmentJSON(a), sessions})
}

type ControlResult struct {
	Attachment Attachment
	Err        error
}

// A cancelled request is discarded before execution. The buffered reply cannot
// block the UE if the caller has already timed out.
type ControlRequest struct {
	Context             context.Context
	Action              string
	Gnbs                map[string]*gnb.GNBContext
	Target              string
	ExpectedPDUSessions int
	Generation          uint64
	ExpectedGeneration  uint64
	ExpectedConnection  uint64
	Reply               chan ControlResult
	DidDeregister       bool // read by the scenario only after the UE has exited
	LastGNB             string
}

func (r *ControlRequest) Respond(a Attachment, err error) {
	select {
	case r.Reply <- ControlResult{a, err}:
	default:
	}
}
