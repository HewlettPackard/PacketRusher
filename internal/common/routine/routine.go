package routine

import (
	"sync"
)

type Routine[T any] struct {
	routineRunning bool
	ch             chan T
	mutex          sync.Mutex
}

func NewRoutine[T any]() Routine[T] {
	return Routine[T]{
		routineRunning: true,
		ch:             make(chan T),
	}
}

func (r *Routine[T]) DoIfRunning(f func(chan<- T)) {
	r.mutex.Lock()
	if r.routineRunning {
		f(r.ch)
	}
	r.mutex.Unlock()
}

func (r *Routine[T]) SendIfRunning(t T) {
	r.mutex.Lock()
	if r.routineRunning {
		r.ch <- t
	}
	r.mutex.Unlock()
}

func (r *Routine[T]) Read() <-chan T {
	return r.ch
}

func (r *Routine[T]) StopRunning() {
	r.mutex.Lock()

	r.routineRunning = false
	select {
	case <-r.ch:
	default:
	}

	r.mutex.Unlock()
}
