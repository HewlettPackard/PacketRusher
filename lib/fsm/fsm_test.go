package fsm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Opened StateType = "Opened"
	Closed StateType = "Closed"
)

const (
	Open  EventType = "Open"
	Close EventType = "Close"
)

func TestState(t *testing.T) {
	s := NewState(Closed)

	assert.Equal(t, Closed, s.Current(), "Current() failed")
	assert.True(t, s.Is(Closed), "Is() failed")

	s.Set(Opened)

	assert.Equal(t, Opened, s.Current(), "Current() failed")
	assert.True(t, s.Is(Opened), "Is() failed")
}

func TestFSM(t *testing.T) {
	f, err := NewFSM(
		Transitions{
			{Event: Open, From: Closed, To: Opened},
			{Event: Close, From: Opened, To: Closed},
			{Event: Open, From: Opened, To: Opened},
			{Event: Close, From: Closed, To: Closed},
		},
		Callbacks{
			Opened: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
			Closed: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
		},
	)

	s := NewState(Closed)

	assert.Nil(t, err, "NewFSM() failed")

	assert.Nil(t, f.SendEvent(s, Open, ArgsType{"TestArg": "test arg"}), "SendEvent() failed")
	assert.Nil(t, f.SendEvent(s, Close, ArgsType{"TestArg": "test arg"}), "SendEvent() failed")
	assert.True(t, s.Is(Closed), "Transition failed")

	fakeEvent := EventType("fake event")
	assert.EqualError(t, f.SendEvent(s, fakeEvent, nil),
		fmt.Sprintf("Unknown transition[From: %s, Event: %s]", s.Current(), fakeEvent))
}

func TestFSMInitFail(t *testing.T) {
	duplicateTrans := Transition{
		Event: Close, From: Opened, To: Closed,
	}
	_, err := NewFSM(
		Transitions{
			{Event: Open, From: Closed, To: Opened},
			duplicateTrans,
			duplicateTrans,
			{Event: Open, From: Opened, To: Opened},
			{Event: Close, From: Closed, To: Closed},
		},
		Callbacks{
			Opened: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
			Closed: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
		},
	)

	assert.EqualError(t, err, fmt.Sprintf("Duplicate transition: %+v", duplicateTrans))

	fakeState := StateType("fake state")

	_, err = NewFSM(
		Transitions{
			{Event: Open, From: Closed, To: Opened},
			{Event: Close, From: Opened, To: Closed},
			{Event: Open, From: Opened, To: Opened},
			{Event: Close, From: Closed, To: Closed},
		},
		Callbacks{
			Opened: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
			Closed: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
			fakeState: func(state *State, event EventType, args ArgsType) {
				fmt.Printf("event [%+v] at state [%+v]\n", event, state.Current())
			},
		},
	)

	assert.EqualError(t, err, fmt.Sprintf("Unknown state: %+v", fakeState))
}
