/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package control drives the UEs of a running test through a local Unix socket:
// a connection carries one JSON request and its JSON response.
package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"my5G-RANTester/internal/common/tools"
	gnbContext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Request is an action on a UE. It is also a step of a JSON scenario.
type Request struct {
	UeId      int    `json:"ue"`
	Action    string `json:"action"`
	Target    string `json:"target,omitempty"`
	TimeoutMs int    `json:"timeout_ms,omitempty"`
}

type Gnb struct {
	Id    string `json:"id"`
	Ready bool   `json:"ready"`
}

type Response struct {
	Ues   []procedures.UeStatus `json:"ues"`
	Gnbs  []Gnb                 `json:"gnbs"`
	Error string                `json:"error,omitempty"`
}

// Validate refuses a request that no UE could run, whatever its state.
func (request Request) Validate() error {
	handover := false
	switch request.Action {
	case "inspect", "wait", "register", "deregister", "idle", "reconnect":
	case "xn-handover", "ng-handover":
		handover = true
	default:
		return fmt.Errorf("unknown action %q", request.Action)
	}
	switch {
	case request.UeId < 0 || request.UeId == 0 && request.Action != "inspect":
		return fmt.Errorf("%s requires the ID of a UE, starting from 1", request.Action)
	case handover != (request.Target != ""):
		return errors.New("a target gNB ID is required by handovers, and only by them")
	case request.TimeoutMs < 0:
		return errors.New("the timeout cannot be negative")
	}
	return nil
}

func (request Request) timeout() time.Duration {
	if request.TimeoutMs == 0 {
		return 30 * time.Second
	}
	return time.Duration(request.TimeoutMs) * time.Millisecond
}

// done reports whether a UE in this status has completed the requested action.
func (request Request) done(status procedures.UeStatus) bool {
	switch request.Action {
	case "inspect":
		return true
	case "deregister":
		return status.State == "parked"
	case "idle":
		return status.State == "idle" && !status.Connected
	case "xn-handover", "ng-handover":
		return status.Ready && status.GnbId == request.Target
	}
	return status.Ready
}

// Server answers the control requests of a test. Closing it removes its socket.
type Server struct {
	net.Listener
	gnbs map[string]*gnbContext.GNBContext
	lock sync.Mutex
	ues  []*tools.UESimulation // by UE ID, nil until the UE is started
}

// Listen serves the control requests on a Unix socket reserved to its owner.
func Listen(socket string, gnbs map[string]*gnbContext.GNBContext, numUes int) (*Server, error) {
	// A socket left behind by a crash would prevent listening, unlike one in use.
	if conn, err := net.Dial("unix", socket); err == nil {
		conn.Close()
		return nil, fmt.Errorf("control socket %s is already in use", socket)
	}
	if info, err := os.Lstat(socket); err == nil && info.Mode()&os.ModeSocket != 0 {
		os.Remove(socket)
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(socket, 0600); err != nil {
		listener.Close()
		return nil, err
	}
	server := &Server{Listener: listener, gnbs: gnbs, ues: make([]*tools.UESimulation, numUes)}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handle(conn)
		}
	}()
	return server, nil
}

// AddUe makes a started UE reachable by the control requests.
func (server *Server) AddUe(ueId int, simulation *tools.UESimulation) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.ues[ueId-1] = simulation
}

func (server *Server) handle(conn net.Conn) {
	defer conn.Close()
	var request Request
	var response Response
	err := json.NewDecoder(io.LimitReader(conn, 4096)).Decode(&request)
	if err == nil {
		err = request.Validate()
	}
	if err == nil {
		response, err = server.run(request)
	}
	if err != nil {
		response.Error = err.Error()
	}
	json.NewEncoder(conn).Encode(response)
}

// run starts the action, then inspects its UE until the action is done.
func (server *Server) run(request Request) (Response, error) {
	deadline := time.Now().Add(request.timeout())
	action := request.Action
	if action == "wait" {
		action = "inspect"
	}
	response, err := server.control(request.UeId, action, request.Target)
	for err == nil && request.UeId != 0 && !request.done(response.Ues[0]) {
		if time.Now().After(deadline) {
			return response, fmt.Errorf("UE %d: %s timed out in state %s", request.UeId, request.Action, response.Ues[0].State)
		}
		time.Sleep(100 * time.Millisecond)
		response, err = server.control(request.UeId, "inspect", "")
	}
	return response, err
}

// control runs an action on a UE, or on every UE when ueId is 0.
func (server *Server) control(ueId int, action, target string) (Response, error) {
	response := Response{Ues: []procedures.UeStatus{}, Gnbs: []Gnb{}}
	server.lock.Lock()
	ues := append([]*tools.UESimulation{}, server.ues...)
	server.lock.Unlock()
	if ueId > len(ues) {
		return response, fmt.Errorf("unknown UE %d", ueId)
	}
	for i, simulation := range ues {
		if ueId != 0 && ueId != i+1 {
			continue
		}
		status, err := procedures.UeStatus{State: "starting", PduSessions: []int{}}, error(nil)
		if simulation != nil {
			status, err = simulation.Control(action, target)
		} else if action != "inspect" {
			err = errors.New("UE is not started yet")
		}
		status.UeId = i + 1
		response.Ues = append(response.Ues, status)
		if err != nil {
			return response, fmt.Errorf("UE %d: %w", i+1, err)
		}
	}
	for id, gnb := range server.gnbs {
		ready := false
		for amf := range gnb.IterGnbAmf() {
			ready = ready || amf.GetState() == gnbContext.Active
		}
		response.Gnbs = append(response.Gnbs, Gnb{Id: id, Ready: ready})
	}
	sort.Slice(response.Gnbs, func(i, j int) bool { return response.Gnbs[i].Id < response.Gnbs[j].Id })
	return response, nil
}

// Call sends a request to the control socket of a running test.
func Call(socket string, request Request) (Response, error) {
	var response Response
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return response, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(request.timeout() + 5*time.Second))
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return response, err
	}
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return response, err
	}
	if response.Error != "" {
		return response, errors.New(response.Error)
	}
	return response, nil
}

// RunScenario runs the steps of a JSON scenario in order, each one once the
// previous one is done, and stops at the first failure.
func RunScenario(socket, path string) error {
	var scenario struct {
		Steps []Request `json:"steps"`
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scenario); err != nil {
		return fmt.Errorf("invalid scenario %s: %w", path, err)
	}
	if len(scenario.Steps) == 0 {
		return fmt.Errorf("scenario %s has no steps", path)
	}
	for i, step := range scenario.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	for i, step := range scenario.Steps {
		if _, err := Call(socket, step); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
		log.Info("[TESTER] Scenario step ", i+1, " done: ", step.Action, " UE ", step.UeId)
	}
	return nil
}
