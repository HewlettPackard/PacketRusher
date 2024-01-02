package fsm

import "sync"

type StateType string

// State is a thread-safe structure that represents the state of a object
// and it can be used for FSM
type State struct {
	// current state of the State object
	current StateType
	// stateMutex ensures that all operations to current is thread-safe
	stateMutex sync.RWMutex
}

// NewState create a State object with current state set to initState
func NewState(initState StateType) *State {
	s := &State{current: initState}
	return s
}

// Current get the current state
func (state *State) Current() StateType {
	state.stateMutex.RLock()
	defer state.stateMutex.RUnlock()
	return state.current
}

// Is return true if the current state is equal to target
func (state *State) Is(target StateType) bool {
	state.stateMutex.RLock()
	defer state.stateMutex.RUnlock()
	return state.current == target
}

// Set current state to next
func (state *State) Set(next StateType) {
	state.stateMutex.Lock()
	defer state.stateMutex.Unlock()
	state.current = next
}
