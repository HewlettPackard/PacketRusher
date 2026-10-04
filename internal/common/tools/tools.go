/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package tools

import (
	"context"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	gnbCxt "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	ueCtx "my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"errors"

	log "github.com/sirupsen/logrus"
)

// CreateGnbs starts count gNBs. If one of them fails to start, those already created
// are terminated and the error is returned.
func CreateGnbs(ctx context.Context, count int, cfg config.Config, wg *sync.WaitGroup) (map[string]*gnbCxt.GNBContext, error) {
	gnbs := make(map[string]*gnbCxt.GNBContext)
	// Each gNB have their own IP address on both N2 and N3
	// TODO: Limitation for now, these IPs must be sequential, eg:
	// gnb[0].n2_ip = 192.168.2.10, gnb[0].n3_ip = 192.168.3.10
	// gnb[1].n2_ip = 192.168.2.11, gnb[1].n3_ip = 192.168.3.11
	// ...
	basePLMN := cfg.GNodeB.PlmnList
	if _, err := basePLMN.GNBIDAt(count - 1); err != nil {
		return nil, fmt.Errorf("gNB identifier range: %w", err)
	}
	cfg.GNodeB.PlmnList.GnbId, _ = basePLMN.GNBIDAt(0)
	for i := 1; i <= count; i++ {
		created, err := gnb.InitGnb(ctx, cfg, wg)
		if err != nil {
			for _, started := range gnbs {
				started.Terminate()
			}
			return nil, err
		}
		gnbs[cfg.GNodeB.PlmnList.GnbId] = created
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the incremented IPs, still belong to the interfaces' subnet
		if i < count {
			cfg.GNodeB.PlmnList.GnbId, _ = basePLMN.GNBIDAt(i)
		}
		// Setup retries may have consumed addresses beyond the configured start.
		// Allocate after the settled endpoints, preserving the configured ports.
		cfg.GNodeB.ControlIF.AddrPort = netip.AddrPortFrom(created.GetGnbIpPort().Addr().Next(), cfg.GNodeB.ControlIF.Port())
		cfg.GNodeB.DataIF.AddrPort = netip.AddrPortFrom(created.GetN3GnbIp().Next(), cfg.GNodeB.DataIF.Port())
	}
	return gnbs, nil
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

type UESimulationConfig struct {
	UeId                     int
	Gnbs                     map[string]*gnbCxt.GNBContext
	Cfg                      config.Config
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

// Control runs a control socket action on the UE and reports the UE's status.
func (simulation *UESimulation) Control(action, target string) (procedures.UeStatus, error) {
	request := &procedures.ControlRequest{Action: action, Target: target, Reply: make(chan procedures.ControlReply, 1)}
	reply := noUe(request, "stopped")
	if simulation.Send(procedures.UeTesterMessage{Type: procedures.Control, Control: request}) {
		select {
		case reply = <-request.Reply:
		case <-simulation.done:
		}
	}
	return reply.Status, reply.Err
}

// noUe answers a control request the scenario has no UE to run on: it can still
// be inspected, in the given state.
func noUe(request *procedures.ControlRequest, state string) procedures.ControlReply {
	reply := procedures.ControlReply{Status: procedures.UeStatus{State: state, PduSessions: []int{}}}
	if request.Action != "inspect" {
		reply.Err = fmt.Errorf("cannot %s a UE that is %s", request.Action, state)
	}
	return reply
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
		stopping := false
		// A control deregistration parks the scenario: its UE stays deregistered
		// until a control registration, whatever the registration loop says.
		parked := false
		for iteration := 1; ; iteration++ {
			wg.Add(1)
			ueRx := make(chan procedures.UeTesterMessage)
			ueTx := ue.NewUE(ueCfg, simConfig.UeId, ueRx, simConfig.Gnbs[simConfig.gnbID(0)].GetInboundChannel(), wg)
			pending := []procedures.UeTesterMessage{{Type: procedures.Registration}}

			after := func(milliseconds int) <-chan time.Time {
				if milliseconds <= 0 {
					return nil
				}
				return time.After(time.Duration(milliseconds) * time.Millisecond)
			}
			deregistrationChannel := after(simConfig.TimeBeforeDeregistration)
			ngapHandoverChannel := after(simConfig.TimeBeforeNgapHandover)
			xnHandoverChannel := after(simConfig.TimeBeforeXnHandover)
			idleChannel := after(simConfig.TimeBeforeIdle)
			var reconnectChannel <-chan time.Time
			nextHandoverId := 0
			registered := false
			state := ueCtx.MM5G_NULL
			alive := true
			// The UE takes no command after Terminate or Kill, so the scenario
			// answers the control requests still queued for it.
			dropPending := func() {
				for _, message := range pending {
					if message.Control != nil {
						message.Control.Reply <- noUe(message.Control, "deregistering")
					}
				}
				pending = nil
			}
			endIteration := func() {
				if ueRx != nil && !stopping {
					pending = append(pending, procedures.UeTesterMessage{Type: procedures.Terminate})
				}
			}
			// control runs a control socket action on the status its UE just reported.
			control := func(request *procedures.ControlRequest, status procedures.UeStatus) {
				var source *gnbCxt.GNBContext
				for id, gnb := range simConfig.Gnbs {
					if gnb.GetInboundChannel() == status.GnbInboundChannel {
						status.GnbId, source = id, gnb
					}
				}
				status.Ready = status.Ready && len(status.PduSessions) >= simConfig.NumPduSessions
				target := simConfig.Gnbs[request.Target]
				var err error
				switch action := request.Action; {
				case action == "inspect":
				case ueRx == nil || stopping:
					err = errors.New("UE is terminating")
				case action == "deregister" && status.State == "registered":
					parked = true
					endIteration()
				case action == "idle" && status.Ready:
					pending = append(pending, procedures.UeTesterMessage{Type: procedures.Idle})
				case action == "reconnect" && status.State == "idle" && !status.Connected:
					pending = append(pending, procedures.UeTesterMessage{Type: procedures.ServiceRequest})
				case (action == "xn-handover" || action == "ng-handover") && (target == nil || target == source):
					err = fmt.Errorf("target gNB %q is unknown or already serves the UE", request.Target)
				case (action == "xn-handover" || action == "ng-handover") && len(status.PduSessions) == 0:
					err = errors.New("a UE without PDU session cannot be handed over")
				case action == "xn-handover" && status.Ready:
					trigger.TriggerXnHandover(source, target, int64(simConfig.UeId))
				case action == "ng-handover" && status.Ready:
					trigger.TriggerNgapHandover(source, target, int64(simConfig.UeId))
				default:
					err = fmt.Errorf("cannot %s a UE that is %s (ready: %t)", action, status.State, status.Ready)
				}
				request.Reply <- procedures.ControlReply{Status: status, Err: err}
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
				var commands <-chan procedures.UeTesterMessage
				if len(pending) < 32 {
					commands = simulation.commands
				}
				select {
				case commandTx <- command:
					pending = pending[1:]
					if command.Type == procedures.Terminate || command.Type == procedures.Kill {
						ueRx = nil
						dropPending()
					}
				case <-deregistrationChannel:
					deregistrationChannel = nil
					endIteration()
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
					if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
						stopping = true
					}
					if ueRx != nil {
						pending = append(pending, msg)
					} else if msg.Control != nil {
						msg.Control.Reply <- noUe(msg.Control, "deregistering")
					}
				case msg, open := <-ueTx:
					if !open {
						alive = false
						continue
					}
					if msg.Control != nil {
						control(msg.Control, msg.Status)
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
			dropPending()
			if stopping || !parked && (!simConfig.RegistrationLoop || (simConfig.LoopCount != 0 && iteration >= simConfig.LoopCount)) {
				return
			}
			betweenRegistrations := "deregistered"
			if parked {
				betweenRegistrations = "parked"
			}
			// Global shutdown remains receivable between registration attempts.
			restart := time.NewTimer(time.Duration(simConfig.TimeBeforeReregistration) * time.Millisecond)
			waiting := true
			for waiting {
				select {
				case <-restart.C:
					waiting = parked
				case msg := <-simulation.commands:
					if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
						restart.Stop()
						return
					}
					if request := msg.Control; request != nil {
						reply := noUe(request, betweenRegistrations)
						if parked && request.Action == "register" {
							parked, waiting, reply.Err = false, false, nil
						}
						request.Reply <- reply
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
