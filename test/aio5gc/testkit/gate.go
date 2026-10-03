// SPDX-License-Identifier: Apache-2.0
package testkit

import (
	"context"
	"sync"
)

// Gate pauses a scenario phase without blocking core shutdown. Reached closes
// once a real decoded message/ FSM entry reaches the phase; Open releases all
// callers. A gate does not synthesize protocol success or alter wire messages.
type Gate struct {
	reached chan struct{}
	opened  chan struct{}
	reach   sync.Once
	open    sync.Once
}

func NewGate() *Gate                     { return &Gate{reached: make(chan struct{}), opened: make(chan struct{})} }
func (g *Gate) Reached() <-chan struct{} { return g.reached }
func (g *Gate) Open()                    { g.open.Do(func() { close(g.opened) }) }
func (g *Gate) Wait(ctx context.Context) error {
	g.reach.Do(func() { close(g.reached) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	select {
	case <-g.opened:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
