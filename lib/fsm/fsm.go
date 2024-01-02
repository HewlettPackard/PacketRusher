package fsm

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type (
	EventType string
	ArgsType  map[string]interface{}
)

const (
	ArgCallerInfo string = "CallerInfo"
)

type (
	Callback  func(*State, EventType, ArgsType)
	Callbacks map[StateType]Callback
)

// Transition defines a transition
// that a Event is triggered at From state,
// and transfer to To state after the Event
type Transition struct {
	Event EventType
	From  StateType
	To    StateType
}

type Transitions []Transition

type eventKey struct {
	Event EventType
	From  StateType
}

// Entry/Exit event are defined by fsm package
const (
	EntryEvent EventType = "Entry event"
	ExitEvent  EventType = "Exit event"
)

type FSM struct {
	// transitions stores one transition for each event
	transitions map[eventKey]Transition
	// callbacks stores one callback function for one state
	callbacks map[StateType]Callback
}

// NewFSM create a new FSM object then registers transitions and callbacks to it
func NewFSM(transitions Transitions, callbacks Callbacks) (*FSM, error) {
	fsm := &FSM{
		transitions: make(map[eventKey]Transition),
		callbacks:   make(map[StateType]Callback),
	}

	allStates := make(map[StateType]bool)

	for _, transition := range transitions {
		key := eventKey{
			Event: transition.Event,
			From:  transition.From,
		}
		if _, ok := fsm.transitions[key]; ok {
			return nil, fmt.Errorf("Duplicate transition: %+v", transition)
		} else {
			fsm.transitions[key] = transition
			allStates[transition.From] = true
			allStates[transition.To] = true
		}
	}

	for state, callback := range callbacks {
		if _, ok := allStates[state]; !ok {
			return nil, fmt.Errorf("Unknown state: %+v", state)
		} else {
			fsm.callbacks[state] = callback
		}
	}
	return fsm, nil
}

// SendEvent triggers a callback with an event, and do transition after callback if need
// There are 3 types of callback:
//   - on exit callback: call when fsm leave one state, with ExitEvent event
//   - event callback: call when user trigger a user-defined event
//   - on entry callback: call when fsm enter one state, with EntryEvent event
func (fsm *FSM) SendEvent(state *State, event EventType, args ArgsType) error {
	key := eventKey{
		From:  state.Current(),
		Event: event,
	}

	if trans, ok := fsm.transitions[key]; ok {
		callerInfo := ""
		if argCallerInfo, ok2 := args[ArgCallerInfo]; ok2 {
			callerInfo = fmt.Sprintf("[%s] ", argCallerInfo.(string))
		}

		log.Infof("%sHandle event[%s], transition from [%s] to [%s]",
			callerInfo, event, trans.From, trans.To)

		// event callback
		fsm.callbacks[trans.From](state, event, args)

		// exit callback
		if trans.From != trans.To {
			fsm.callbacks[trans.From](state, ExitEvent, args)
		}

		// entry callback
		if trans.From != trans.To {
			state.Set(trans.To)
			fsm.callbacks[trans.To](state, EntryEvent, args)
		}
		return nil
	} else {
		return fmt.Errorf("Unknown transition[From: %s, Event: %s]", state.Current(), event)
	}
}

// ExportDot export fsm in dot format to outfile, which can be visualize by graphviz
func ExportDot(fsm *FSM, outfile string) error {
	dot := `digraph FSM {
	rankdir=LR
	size="100"
    node[width=1 fixedsize=false shape=ellipse style=filled fillcolor="skyblue"]
	`

	for _, trans := range fsm.transitions {
		link := fmt.Sprintf("\t%s -> %s [label=\"%s\"]", trans.From, trans.To, trans.Event)
		dot = dot + "\r\n" + link
	}

	dot = dot + "\r\n}\n"

	if !strings.HasSuffix(outfile, ".dot") {
		outfile = fmt.Sprintf("%s.dot", outfile)
	}

	if file, err := os.Create(outfile); err != nil {
		return err
	} else {
		if _, err = file.WriteString(dot); err != nil {
			return err
		}
		fmt.Printf("Output the FSM to \"%s\"\n", outfile)
		return file.Close()
	}
}
