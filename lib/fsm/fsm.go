package fsm

import (
	"errors"
	"fmt"
)

type State string
type Event string
type HandleFunc func(*FSM, Event, Args) error
type FuncTable map[State]HandleFunc

const (
	EVENT_ENTRY Event = "ENTRY EVENT"
)

type Args map[string]interface{}

type FSM struct {
	state     State
	funcTable FuncTable
}

func NewFuncTable() (table FuncTable) {
	table = make(map[State]HandleFunc)
	return
}

func NewFSM(initState State, table FuncTable) (fsm *FSM, err error) {
	fsm = new(FSM)

	fsm.state = initState
	fsm.funcTable = table
	if _, ok := table[initState]; !ok {
		return nil, errors.New("initial state is not valid")
	}
	err = fsm.funcTable[fsm.state](fsm, EVENT_ENTRY, nil)
	if err != nil {
		return nil, err
	}
	return fsm, nil
}

func (fsm *FSM) Current() State {
	return fsm.state
}

func (fsm *FSM) Check(state State) bool {
	return fsm.state == state
}

func (fsm *FSM) AddState(state State, callback HandleFunc) {
	fsm.funcTable[state] = callback
}

func (fsm *FSM) SendEvent(event Event, args Args) (err error) {
	err = fsm.funcTable[fsm.state](fsm, event, args)
	return
}

/* args is for ENTRY params*/
func (fsm *FSM) Transfer(trans State, args Args) error {
	if _, ok := fsm.funcTable[trans]; !ok {
		return fmt.Errorf("Fsm Tranfer Error : State %s is not exist", trans)
	}
	if trans != fsm.state {
		fsm.state = trans
		err := fsm.funcTable[fsm.state](fsm, EVENT_ENTRY, args)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fsm *FSM) AllStates() (states []State) {
	for key := range fsm.funcTable {
		states = append(states, key)
	}
	return
}

func (fsm *FSM) PrintStates() {
	for key := range fsm.funcTable {
		fmt.Println(key)
	}
}
