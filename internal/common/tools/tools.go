/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	"context"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	gnbCxt "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	ueCtx "my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"strconv"
	"sync"
	"time"

	"errors"

	log "github.com/sirupsen/logrus"
)

func CreateGnbs(count int, cfg config.Config, wg *sync.WaitGroup) map[string]*gnbCxt.GNBContext {
	gnbs := make(map[string]*gnbCxt.GNBContext)
	// Each gNB have their own IP address on both N2 and N3
	// TODO: Limitation for now, these IPs must be sequential, eg:
	// gnb[0].n2_ip = 192.168.2.10, gnb[0].n3_ip = 192.168.3.10
	// gnb[1].n2_ip = 192.168.2.11, gnb[1].n3_ip = 192.168.3.11
	// ...
	if count < 1 {
		log.Fatal("[GNB] At least one gNB is required")
	}
	if _, err := cfg.GNodeB.PlmnList.GNBIDAt(count - 1); err != nil {
		log.Fatalf("[GNB] Invalid identifier range: %v", err)
	}
	cfg.GNodeB.PlmnList.GnbId, _ = cfg.GNodeB.PlmnList.GNBIDAt(0)
	basePLMN := cfg.GNodeB.PlmnList
	for i := 1; i <= count; i++ {
		gnbs[cfg.GNodeB.PlmnList.GnbId] = gnb.InitGnb(cfg, wg)
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the incremented IPs, still belong to the interfaces' subnet
		if i < count {
			cfg.GNodeB.PlmnList.GnbId, _ = basePLMN.GNBIDAt(i)
		}
		cfg.GNodeB.ControlIF = cfg.GNodeB.ControlIF.WithNextAddr()
		cfg.GNodeB.DataIF = cfg.GNodeB.DataIF.WithNextAddr()
	}
	return gnbs
}

func IncrementIP(origIP, cidr string) (string, error) {
	ip := net.ParseIP(origIP)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return origIP, err
	}
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
	if !ipNet.Contains(ip) {
		return origIP, errors.New("Ip is not in provided subnet")
	}
	return ip.String(), nil
}

func gnbIdGenerator(offset int, id string) string {
	value, err := (config.PlmnList{GnbId: id}).GNBIDAt(offset)
	if err != nil {
		log.Fatalf("[GNB] Invalid identifier: %v", err)
	}
	return value
}

type UESimulationConfig struct {
	UeId                     int
	Gnbs                     map[string]*gnbCxt.GNBContext
	Cfg                      config.Config
	ScenarioChan             chan procedures.UeTesterMessage
	TimeBeforeDeregistration int
	TimeBeforeNgapHandover   int
	TimeBeforeXnHandover     int
	TimeBeforeIdle           int
	TimeBeforeReconnecting   int
	NumPduSessions           int
	RegistrationLoop         bool
	LoopCount                int
	TimeBeforeReregistration int
}

// gnbID selects the initial gNB and subsequent handover targets in the same
// round-robin sequence. UE IDs start at 1; gNB offsets start at 0.
func (simConfig UESimulationConfig) gnbID(handoverOffset int) string {
	index := (simConfig.UeId - 1 + handoverOffset) % len(simConfig.Gnbs)
	id, err := simConfig.Cfg.GNodeB.PlmnList.GNBIDAt(index)
	if err != nil {
		log.Fatalf("[GNB] Invalid identifier: %v", err)
	}
	return id
}

// UESimulation tracks the entire scenario, including registration-loop delays.
// Send waits for a live scenario to accept a command or for it to finish.
type UESimulation struct {
	commands    chan procedures.UeTesterMessage
	done        chan struct{}
	controlGate chan struct{}
	config      UESimulationConfig
}

func (simulation *UESimulation) Done() <-chan struct{} { return simulation.done }

func (simulation *UESimulation) Send(message procedures.UeTesterMessage) bool {
	select {
	case <-simulation.done:
		return false
	default:
	}
	select {
	case simulation.commands <- message:
		return true
	case <-simulation.done:
		return false
	}
}

func SimulateSingleUE(simConfig UESimulationConfig, wg *sync.WaitGroup) *UESimulation {
	ueCfg := simConfig.Cfg
	ueCfg.Ue.Msin = IncrementMsin(simConfig.UeId, simConfig.Cfg.Ue.Msin)
	log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", ueCfg.Ue.Msin, " UE")

	simulation := &UESimulation{commands: make(chan procedures.UeTesterMessage), done: make(chan struct{}), config: simConfig, controlGate: make(chan struct{}, 1)}
	// Count the scenario before starting it. It may create further UEs after a
	// registration-loop delay, even when the previous UE has already finished.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(simulation.done)
		scenarioChan := simConfig.ScenarioChan
		stopping := false
		initialGNB := simConfig.gnbID(0)
	iterations:
		for iteration := 1; ; iteration++ {
			wg.Add(1)
			ueRx := make(chan procedures.UeTesterMessage)
			ueTx := ue.NewUE(ueCfg, simConfig.UeId, ueRx, simConfig.Gnbs[initialGNB].GetInboundChannel(), wg)
			pending := []procedures.UeTesterMessage{{Type: procedures.Registration}}

			after := func(milliseconds int) <-chan time.Time {
				if milliseconds == 0 {
					return nil
				}
				return time.After(time.Duration(milliseconds) * time.Millisecond)
			}
			var deregistrationChannel, ngapHandoverChannel, xnHandoverChannel, idleChannel <-chan time.Time
			var reconnectChannel <-chan time.Time
			iterationCtx, cancelIteration := context.WithCancel(context.Background())
			launchControl := func(action, target string) {
				go func() {
					ctx, cancel := context.WithTimeout(iterationCtx, 30*time.Second)
					defer cancel()
					if _, err := simulation.Execute(ctx, action, target); err != nil && !errors.Is(err, context.Canceled) {
						log.Warn("[TESTER] UE ", simConfig.UeId, " ", action, ": ", err)
					}
				}()
			}
			nextHandoverId := 0
			registered := false
			state := ueCtx.MM5G_NULL
			alive := true
			var parkRequest *procedures.ControlRequest
			acceptCommand := func(message procedures.UeTesterMessage) {
				if message.Control != nil {
					r := message.Control
					r.Generation = uint64(iteration)
					if r.Context.Err() != nil {
						r.Respond(procedures.Attachment{}, r.Context.Err())
						return
					}
					if stopping || ueRx == nil {
						r.Respond(procedures.Attachment{}, procedures.ErrStopped)
						return
					}
					if r.ExpectedGeneration != 0 && r.ExpectedGeneration != uint64(iteration) {
						r.Respond(procedures.Attachment{}, procedures.ErrGeneration)
						return
					}
					if r.Action == "deregister" {
						parkRequest = r
					}
				}
				if message.Type == procedures.Terminate || message.Type == procedures.Kill {
					stopping = true
				}
				if ueRx != nil {
					pending = append(pending, message)
				}
			}
			for alive {
				// Queue commands instead of blocking on a send: a UE handling the
				// previous command may first need us to receive its state update.
				var commandTx chan procedures.UeTesterMessage
				var command procedures.UeTesterMessage
				if len(pending) != 0 && ueRx != nil {
					commandTx, command = ueRx, pending[0]
				}
				// Bound command admission while still receiving UE state updates.
				// One registration can schedule at most 16 PDU session commands.
				var commands, legacyCommands <-chan procedures.UeTesterMessage
				if len(pending) < 32 {
					commands, legacyCommands = simulation.commands, scenarioChan
				}
				select {
				case commandTx <- command:
					pending = pending[1:]
					if command.Type == procedures.Terminate || command.Type == procedures.Kill {
						ueRx = nil
						pending = nil
					}
				case <-deregistrationChannel:
					deregistrationChannel = nil
					if ueRx != nil && !stopping {
						pending = append(pending, procedures.UeTesterMessage{Type: procedures.Terminate})
					}
				case <-ngapHandoverChannel:
					ngapHandoverChannel = nil
					if !stopping {
						launchControl("ng-handover", simConfig.gnbID(nextHandoverId+1))
						nextHandoverId++
					}
				case <-xnHandoverChannel:
					xnHandoverChannel = nil
					if !stopping {
						launchControl("xn-handover", simConfig.gnbID(nextHandoverId+1))
						nextHandoverId++
					}
				case <-idleChannel:
					idleChannel = nil
					if ueRx != nil && !stopping {
						launchControl("idle", "")
						reconnectChannel = after(simConfig.TimeBeforeReconnecting)
					}
				case <-reconnectChannel:
					reconnectChannel = nil
					if ueRx != nil && !stopping {
						launchControl("reconnect", "")
					}
				case msg := <-commands:
					acceptCommand(msg)
				case msg, open := <-legacyCommands:
					if !open {
						scenarioChan = nil
						msg = procedures.UeTesterMessage{Type: procedures.Kill}
					}
					acceptCommand(msg)
				case msg, open := <-ueTx:
					if !open {
						alive = false
						continue
					}
					log.Info("[UE] Switched from state ", state, " to state ", msg.StateChange)
					if msg.StateChange == ueCtx.MM5G_REGISTERED && !registered {
						deregistrationChannel = after(simConfig.TimeBeforeDeregistration)
						ngapHandoverChannel = after(simConfig.TimeBeforeNgapHandover)
						xnHandoverChannel = after(simConfig.TimeBeforeXnHandover)
						idleChannel = after(simConfig.TimeBeforeIdle)
						if ueRx != nil && !stopping {
							for session := 0; session < simConfig.NumPduSessions; session++ {
								pending = append(pending, procedures.UeTesterMessage{Type: procedures.NewPDUSession})
							}
						}
						registered = true
					}
					state = msg.StateChange
				}
			}
			cancelIteration()
			for _, msg := range pending {
				if r := msg.Control; r != nil {
					if r.Action == "inspect" && parkRequest != nil && parkRequest.DidDeregister {
						r.Respond(procedures.Attachment{UE: simConfig.UeId, Generation: uint64(iteration), State: "parked", ActivePDUSessions: []uint8{}}, nil)
					} else {
						r.Respond(procedures.Attachment{}, procedures.ErrGeneration)
					}
				}
			}
			if parkRequest != nil && parkRequest.DidDeregister && !stopping {
				if simConfig.Gnbs[parkRequest.LastGNB] != nil {
					initialGNB = parkRequest.LastGNB
				}
				parked := procedures.Attachment{UE: simConfig.UeId, Generation: uint64(iteration), State: "parked", ActivePDUSessions: []uint8{}}
				for {
					select {
					case msg := <-simulation.commands:
						if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
							return
						}
						if r := msg.Control; r != nil {
							if r.Context.Err() != nil {
								r.Respond(parked, r.Context.Err())
								continue
							}
							if r.ExpectedGeneration != 0 && r.ExpectedGeneration != parked.Generation {
								r.Respond(parked, procedures.ErrGeneration)
								continue
							}
							if r.Action == "inspect" {
								r.Respond(parked, nil)
								continue
							}
							if r.Action == "register" {
								parked.Generation++
								parked.State = "starting"
								r.Respond(parked, nil)
								continue iterations
							}
							r.Respond(parked, procedures.ErrNotReady)
						}
					case msg, open := <-scenarioChan:
						if !open || msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
							return
						}
					}
				}
			}
			if stopping || !simConfig.RegistrationLoop || (simConfig.LoopCount != 0 && iteration >= simConfig.LoopCount) {
				return
			}
			// Global shutdown remains receivable between registration attempts.
			restart := time.NewTimer(time.Duration(simConfig.TimeBeforeReregistration) * time.Millisecond)
			waiting := true
			for waiting {
				select {
				case <-restart.C:
					waiting = false
				case msg := <-simulation.commands:
					if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
						restart.Stop()
						return
					}
				case msg, open := <-scenarioChan:
					if !open || msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
						restart.Stop()
						return
					}
				}
			}
		}
	}()
	return simulation
}

func (simulation *UESimulation) request(ctx context.Context, action, target string, generation, connection uint64) (procedures.Attachment, error) {
	r := &procedures.ControlRequest{Context: ctx, Action: action, Target: target, Gnbs: simulation.config.Gnbs,
		ExpectedPDUSessions: simulation.config.NumPduSessions, ExpectedGeneration: generation,
		ExpectedConnection: connection, Reply: make(chan procedures.ControlResult, 1)}
	select {
	case <-ctx.Done():
		return procedures.Attachment{}, ctx.Err()
	case <-simulation.done:
		return procedures.Attachment{}, procedures.ErrStopped
	case simulation.commands <- procedures.UeTesterMessage{Type: procedures.Control, Control: r}:
	}
	select {
	case result := <-r.Reply:
		return result.Attachment, result.Err
	case <-ctx.Done():
		return procedures.Attachment{}, ctx.Err()
	case <-simulation.done:
		return procedures.Attachment{}, procedures.ErrStopped
	}
}

func IncrementMsin(i int, msin string) string {

	msin_int, err := strconv.Atoi(msin)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given MSIN is invalid")
	}
	base := msin_int + (i - 1)

	var imsi string
	if len(msin) == 9 {
		imsi = fmt.Sprintf("%09d", base)
	} else {
		imsi = fmt.Sprintf("%010d", base)
	}
	return imsi
}
