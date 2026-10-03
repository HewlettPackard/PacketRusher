// SPDX-License-Identifier: Apache-2.0
package scenario

import (
	"context"
	"fmt"
	"io"
)

// Steps execute in order. Each action waits for its prerequisite and observed
// completion, so later actions do not depend on registration-duration guesses.
type Plan struct {
	Steps []Request `json:"steps"`
}

func ReadPlan(reader io.Reader) (Plan, error) {
	var p Plan
	if err := Decode(reader, &p); err != nil {
		return p, err
	}
	if len(p.Steps) == 0 || len(p.Steps) > 10000 {
		return p, fmt.Errorf("scenario requires 1 to 10000 steps")
	}
	for i, step := range p.Steps {
		if err := step.Validate(); err != nil {
			return p, fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return p, nil
}

func (p Plan) Run(ctx context.Context, socket string, output io.Writer) error {
	for i, step := range p.Steps {
		response, err := Call(ctx, socket, step)
		if err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
		if _, err := fmt.Fprintf(output, "%d %s UE %d: %s\n", i+1, step.Action, step.UE, stateOf(response)); err != nil {
			return err
		}
	}
	return nil
}

func stateOf(r Response) string {
	if len(r.UEs) == 1 {
		return r.UEs[0].State
	}
	return "inspected"
}
