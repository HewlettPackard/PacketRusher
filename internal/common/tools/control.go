// SPDX-License-Identifier: Apache-2.0
package tools

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"time"

	log "github.com/sirupsen/logrus"
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
	return s.execute(ctx, action, target, 0, 0)
}

// execute also handles the private automatic termination action. Its generation
// is captured when the timer is armed, so a delayed timer cannot end a new UE.
func (s *UESimulation) execute(ctx context.Context, action, target string, generation, connection uint64) (procedures.Attachment, error) {
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
	if generation != 0 && first.Generation != generation || connection != 0 && first.ConnectionGeneration != connection {
		return first, procedures.ErrGeneration
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
	if action == "terminate" {
		// The UE loop ran the legacy release/deregistration path. It is exiting,
		// so inspecting it again would race the next registration iteration.
		return current, nil
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

// terminateWhenReady preserves automatic loop progression if a core never
// accepts a PDU. Cancellation of the owning iteration must not trigger fallback.
// timeout is explicit so tests exercise this production path with a short bound.
func (s *UESimulation) terminateWhenReady(iterationCtx context.Context, generation uint64, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(iterationCtx, timeout)
	defer cancel()
	first, err := s.Inspect(ctx)
	if err != nil {
		return err
	}
	if first.Generation != generation {
		return procedures.ErrGeneration
	}
	_, err = s.execute(ctx, "terminate", "", generation, first.ConnectionGeneration)
	if !errors.Is(err, context.DeadlineExceeded) || iterationCtx.Err() != nil {
		return err
	}
	log.Warn("[TESTER] UE ", s.config.UeId, " automatic deregistration readiness deadline reached; ending generation ", generation)
	// Use a fresh, bounded context because the readiness deadline has expired.
	// The UE actor rechecks both identifiers before running legacy Terminate.
	cleanupCtx, cancelCleanup := context.WithTimeout(iterationCtx, 10*time.Second)
	defer cancelCleanup()
	if _, fallbackErr := s.request(cleanupCtx, "terminate-after-timeout", "", generation, first.ConnectionGeneration); fallbackErr != nil {
		return fallbackErr
	}
	return nil
}
