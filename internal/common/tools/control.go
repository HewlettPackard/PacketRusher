// SPDX-License-Identifier: Apache-2.0
package tools

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"time"
)

func (s *UESimulation) Inspect(ctx context.Context) (procedures.Attachment, error) {
	return s.request(ctx, "inspect", "", 0, 0)
}

// Execute waits for observed state, rechecks it on the UE loop, and waits for
// local completion. One UE's actions are serialized; inspection remains live.
func (s *UESimulation) Execute(ctx context.Context, action, target string) (procedures.Attachment, error) {
	switch action {
	case "wait", "register", "deregister", "idle", "reconnect", "xn-handover", "ng-handover":
	default:
		return procedures.Attachment{}, fmt.Errorf("unknown control action %q", action)
	}
	select {
	case s.controlGate <- struct{}{}:
		defer func() { <-s.controlGate }()
	case <-ctx.Done():
		return procedures.Attachment{}, ctx.Err()
	case <-s.done:
		return procedures.Attachment{}, procedures.ErrStopped
	}
	first, err := s.Inspect(ctx)
	if err != nil {
		return first, err
	}
	if action == "xn-handover" || action == "ng-handover" {
		if s.config.Gnbs[target] == nil || target == first.GNB {
			return first, fmt.Errorf("handover target must be another configured gNB")
		}
	} else if target != "" {
		return first, fmt.Errorf("--target applies only to handover")
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	current := first
	for {
		if current.Generation != first.Generation || current.ConnectionGeneration != first.ConnectionGeneration {
			return current, procedures.ErrGeneration
		}
		ready := current.Ready
		if action == "deregister" {
			ready = current.State == "registered" && current.Connected
		}
		if action == "register" {
			ready = current.State == "parked"
		}
		if action == "reconnect" {
			ready = current.State == "idle" && !current.Connected
		}
		if action == "wait" && ready {
			return current, nil
		}
		if action != "wait" && ready {
			result, requestErr := s.request(ctx, action, target, first.Generation, first.ConnectionGeneration)
			if requestErr == nil {
				current = result
				break
			}
			if !errors.Is(requestErr, procedures.ErrNotReady) {
				return result, requestErr
			}
		}
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		case <-ticker.C:
		}
		current, err = s.Inspect(ctx)
		if err != nil {
			return current, err
		}
	}
	expectedGeneration := first.Generation
	if action == "register" {
		expectedGeneration++
	}
	for {
		current, err = s.Inspect(ctx)
		if err != nil {
			return current, err
		}
		if current.Generation != expectedGeneration {
			return current, procedures.ErrGeneration
		}
		complete := current.Ready
		switch action {
		case "deregister":
			complete = current.State == "parked"
		case "idle":
			complete = current.State == "idle" && !current.Connected
		case "xn-handover", "ng-handover":
			complete = complete && current.GNB == target
		}
		if complete {
			return current, nil
		}
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		case <-ticker.C:
		}
	}
}
