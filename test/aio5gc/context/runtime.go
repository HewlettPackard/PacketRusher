// SPDX-License-Identifier: Apache-2.0
package context

import (
	stdcontext "context"
	"errors"
	"sync"
	"time"
)

// runtime owns admission, cancellation and all workers/resources of one core.
// Worker enrollment and shutdown share a lock: Wait never races a new Add.
type runtime struct {
	mu        sync.Mutex
	ctx       stdcontext.Context
	cancel    stdcontext.CancelFunc
	stopping  bool
	closed    chan struct{}
	resources map[*resource]struct{}
	workers   sync.WaitGroup
	closeErr  error
}
type resource struct {
	once  sync.Once
	close func() error
	err   error
}

func (r *resource) release() error { r.once.Do(func() { r.err = r.close() }); return r.err }
func (r *runtime) initLocked() {
	if r.ctx == nil {
		r.ctx, r.cancel = stdcontext.WithCancel(stdcontext.Background())
		r.closed = make(chan struct{})
		r.resources = make(map[*resource]struct{})
	}
}

// Context is canceled before any listener or association is closed. Scenario
// hooks must honor it (testkit.Gate does) so inline dispatch can join shutdown.
func (a *Aio5gc) Context() stdcontext.Context {
	a.runtime.mu.Lock()
	defer a.runtime.mu.Unlock()
	a.runtime.initLocked()
	return a.runtime.ctx
}

// Own registers a resource and returns an idempotent retirement function. An
// accepted connection retires immediately when its reader ends, rather than
// being retained in the core until the entire test finishes.
func (a *Aio5gc) Own(close func() error) func() error {
	r := &resource{close: close}
	a.runtime.mu.Lock()
	a.runtime.initLocked()
	if a.runtime.stopping {
		a.runtime.mu.Unlock()
		_ = r.release()
		return r.release
	}
	a.runtime.resources[r] = struct{}{}
	a.runtime.mu.Unlock()
	return func() error {
		err := r.release()
		a.runtime.mu.Lock()
		delete(a.runtime.resources, r)
		a.runtime.mu.Unlock()
		return err
	}
}
func (a *Aio5gc) RegisterCloser(close func() error) { a.Own(close) }
func (a *Aio5gc) BeginWork() bool {
	a.runtime.mu.Lock()
	defer a.runtime.mu.Unlock()
	a.runtime.initLocked()
	if a.runtime.stopping {
		return false
	}
	a.runtime.workers.Add(1)
	return true
}
func (a *Aio5gc) EndWork() { a.runtime.workers.Done() }

// Go enrolls before spawning, so an admitted goroutine cannot escape Close.
func (a *Aio5gc) Go(work func()) bool {
	if !a.BeginWork() {
		return false
	}
	go func() { defer a.EndWork(); work() }()
	return true
}

// Stop is nonblocking. CloseContext also waits for every admitted worker and
// closer; an expired caller deadline never claims that shutdown completed.
func (a *Aio5gc) Stop() {
	r := &a.runtime
	r.mu.Lock()
	r.initLocked()
	if r.stopping {
		r.mu.Unlock()
		return
	}
	r.stopping = true
	r.cancel()
	resources := r.resources
	r.resources = make(map[*resource]struct{})
	r.mu.Unlock()
	go func() {
		var errs []error
		for resource := range resources {
			if err := resource.release(); err != nil {
				errs = append(errs, err)
			}
		}
		r.workers.Wait()
		r.mu.Lock()
		r.closeErr = errors.Join(errs...)
		r.mu.Unlock()
		close(r.closed)
	}()
}
func (a *Aio5gc) CloseContext(ctx stdcontext.Context) error {
	a.Stop()
	select {
	case <-a.runtime.closed:
		a.runtime.mu.Lock()
		defer a.runtime.mu.Unlock()
		return a.runtime.closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (a *Aio5gc) Close() error {
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 10*time.Second)
	defer cancel()
	return a.CloseContext(ctx)
}
func (a *Aio5gc) ResourceCount() int {
	a.runtime.mu.Lock()
	defer a.runtime.mu.Unlock()
	return len(a.runtime.resources)
}
