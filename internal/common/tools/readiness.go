// SPDX-License-Identifier: Apache-2.0
package tools

import (
	"context"
	"fmt"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"sort"
	"time"
)

func WaitGnbs(ctx context.Context, gnbs map[string]*gnb.GNBContext) error {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		var pending []string
		for id, node := range gnbs {
			if !node.NGSetupReady() {
				pending = append(pending, id)
			}
		}
		if len(pending) == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			sort.Strings(pending)
			return fmt.Errorf("waiting for NG Setup responses from %v: %w", pending, ctx.Err())
		case <-ticker.C:
		}
	}
}
