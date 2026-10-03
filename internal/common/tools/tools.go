/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	gnbCxt "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
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
	baseGnbId := cfg.GNodeB.PlmnList.GnbId
	for i := 1; i <= count; i++ {
		gnbs[cfg.GNodeB.PlmnList.GnbId] = gnb.InitGnb(cfg, wg)
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the incremented IPs, still belong to the interfaces' subnet
		cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i, baseGnbId)
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

func gnbIdGenerator(i int, gnbId string) string {

	gnbId_int, err := strconv.ParseInt(gnbId, 16, 0)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given gnbId is invalid")
	}
	base := int(gnbId_int) + i

	gnbId = fmt.Sprintf("%06X", base)
	return gnbId
}

type UESimulationConfig struct {
	UeId                     int
	Gnbs                     map[string]*gnbCxt.GNBContext
	Cfg                      config.Config
	ScenarioChan             chan procedures.UeTesterMessage
	TimeBeforeDeregistration int
	// DeregistrationTrigger optionally replaces the wall-clock timer. Each value
	// gracefully ends the current iteration, without stopping registration loops.
	// Closing it disables explicit triggers; scenario shutdown remains receivable.
	// A nil channel preserves the normal TimeBeforeDeregistration behavior.
	DeregistrationTrigger    <-chan struct{}
	TimeBeforeNgapHandover   int
	TimeBeforeXnHandover     int
	TimeBeforeIdle           int
	TimeBeforeReconnecting   int
	NumPduSessions           int
	RegistrationLoop         bool
	LoopCount                int
	TimeBeforeReregistration int
}

// UESimulation tracks the entire scenario, including registration-loop delays.
// Send waits for a live scenario to accept a command or for it to finish.
type UESimulation struct {
	commands chan procedures.UeTesterMessage
	done     chan struct{}
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

// gnbID selects the initial gNB and subsequent handover targets in the same
// round-robin sequence. UE IDs start at 1; gNB offsets start at 0.
func (simConfig UESimulationConfig) gnbID(handoverOffset int) string {
	index := (simConfig.UeId - 1 + handoverOffset) % len(simConfig.Gnbs)
	if index == 0 {
		// CreateGnbs preserves the configured ID as the first map key. Hex
		// letter case must therefore be retained when returning to that gNB.
		return simConfig.Cfg.GNodeB.PlmnList.GnbId
	}
	return gnbIdGenerator(index, simConfig.Cfg.GNodeB.PlmnList.GnbId)
}

func SimulateSingleUE(simConfig UESimulationConfig, wg *sync.WaitGroup) *UESimulation {
	ueCfg := simConfig.Cfg
	ueCfg.Ue.Msin = IncrementMsin(simConfig.UeId, simConfig.Cfg.Ue.Msin)
	log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", ueCfg.Ue.Msin, " UE")

	simulation := &UESimulation{commands: make(chan procedures.UeTesterMessage), done: make(chan struct{})}
	// Count the scenario before starting it. It may create further UEs after a
	// registration-loop delay, even when the previous UE has already finished.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(simulation.done)
		scenarioChan := simConfig.ScenarioChan
		deregistrationTrigger := simConfig.DeregistrationTrigger
		stopping := false
		for iteration := 1; ; iteration++ {
			wg.Add(1)
			ueRx := make(chan procedures.UeTesterMessage)
			ueTx := ue.NewUE(ueCfg, simConfig.UeId, ueRx, simConfig.Gnbs[simConfig.gnbID(0)].GetInboundChannel(), wg)
			pending := []procedures.UeTesterMessage{{Type: procedures.Registration}}

			after := func(milliseconds int) <-chan time.Time {
				if milliseconds == 0 {
					return nil
				}
				return time.After(time.Duration(milliseconds) * time.Millisecond)
			}
			var deregistrationChannel <-chan time.Time
			iterationTrigger := deregistrationTrigger
			if simConfig.DeregistrationTrigger == nil {
				deregistrationChannel = after(simConfig.TimeBeforeDeregistration)
			}
			ngapHandoverChannel := after(simConfig.TimeBeforeNgapHandover)
			xnHandoverChannel := after(simConfig.TimeBeforeXnHandover)
			idleChannel := after(simConfig.TimeBeforeIdle)
			var reconnectChannel <-chan time.Time
			nextHandoverId := 0
			registered := false
			state := ueCtx.MM5G_NULL
			alive := true
			acceptCommand := func(message procedures.UeTesterMessage) {
				if message.Type == procedures.Terminate || message.Type == procedures.Kill {
					stopping = true
				}
				if ueRx != nil {
					pending = append(pending, message)
				}
			}
			endIteration := func() {
				if ueRx != nil && !stopping {
					pending = append(pending, procedures.UeTesterMessage{Type: procedures.Terminate})
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
					endIteration()
				case _, open := <-iterationTrigger:
					iterationTrigger = nil // At most one graceful stop per iteration.
					if open {
						endIteration()
					} else {
						// Do not re-arm a closed channel for subsequent iterations.
						deregistrationTrigger = nil
					}
				case <-ngapHandoverChannel:
					ngapHandoverChannel = nil
					if !stopping {
						trigger.TriggerNgapHandover(simConfig.Gnbs[simConfig.gnbID(nextHandoverId)], simConfig.Gnbs[simConfig.gnbID(nextHandoverId+1)], int64(simConfig.UeId))
						nextHandoverId++
					}
				case <-xnHandoverChannel:
					xnHandoverChannel = nil
					if !stopping {
						trigger.TriggerXnHandover(simConfig.Gnbs[simConfig.gnbID(nextHandoverId)], simConfig.Gnbs[simConfig.gnbID(nextHandoverId+1)], int64(simConfig.UeId))
						nextHandoverId++
					}
				case <-idleChannel:
					idleChannel = nil
					if ueRx != nil && !stopping {
						pending = append(pending, procedures.UeTesterMessage{Type: procedures.Idle})
						reconnectChannel = after(simConfig.TimeBeforeReconnecting)
					}
				case <-reconnectChannel:
					reconnectChannel = nil
					if ueRx != nil && !stopping {
						pending = append(pending, procedures.UeTesterMessage{Type: procedures.ServiceRequest})
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
