/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	ueCtx "my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Runner struct {
	gnbs map[string]*context.GNBContext
	ues  map[int]*ueSim
}

type ueSim struct {
	ongoingTask bool
	nextTasks   []Task
	taskChan    chan Task
	reportChan  chan report
}

type UEScenario struct {
	Config config.Ue
	Tasks  []Task
}

type report struct {
	success bool
	reason  string
}

func (s *Runner) Start(gnbConfs []config.GNodeB, amfConf config.AMF, ueScenarios []UEScenario, maxRequestRate int) {

	wg := sync.WaitGroup{}

	sender.Init(maxRequestRate)

	s.gnbs = make(map[string]*context.GNBContext)

	for gnbConf := range gnbConfs {
		s.gnbs[gnbConfs[gnbConf].PlmnList.GnbId] = gnb.InitGnb2(gnbConfs[gnbConf], amfConf, &wg) // TODO: replace InitGNB2 with InitGNB
		wg.Add(1)
	}

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(2 * time.Second)

	s.ues = make(map[int]*ueSim)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)
	stopSignal := false
	for !stopSignal {
		done := true
		for ueId := 1; !stopSignal && ueId <= len(ueScenarios); ueId++ {
			if s.ues[ueId] == nil {
				ue := ueSim{
					ongoingTask: false,
					nextTasks:   ueScenarios[ueId-1].Tasks,
					taskChan:    make(chan Task, 1),
					reportChan:  make(chan report, 1),
				}

				go simulateSingleUE(ueId, ueScenarios[ueId-1].Config, ue.taskChan, ue.reportChan, s.gnbs, &wg)

				s.ues[ueId] = &ue
			}

			select {
			case report := <-s.ues[ueId].reportChan:

				if !report.success {
					s.ues[ueId].nextTasks = []Task{}
				}
				s.ues[ueId].ongoingTask = false
			case <-sigStop:
				stopSignal = true
			default:
			}

			if s.ues[ueId].ongoingTask {
				done = false
			} else {
				if len(s.ues[ueId].nextTasks) > 0 {
					s.ues[ueId].taskChan <- s.ues[ueId].nextTasks[0]
					s.ues[ueId].nextTasks = s.ues[ueId].nextTasks[1:]
					s.ues[ueId].ongoingTask = true
					done = false
				}
			}

		}
		if done {
			stopSignal = true
		}
	}
	for _, ue := range s.ues {
		ue.taskChan <- Task{
			TaskType: Terminate,
		}
	}

	time.Sleep(time.Second * 1)
}

func simulateSingleUE(ueId int, ueCfg config.Ue, taskChan chan Task, reportChan chan report, gnbs map[string]*context.GNBContext, wg *sync.WaitGroup) {
	log.Infof("simulation started for ue %d", ueId)

	wg.Add(1)

	ueRx := make(chan procedures.UeTesterMessage)
	var ueTx chan scenario.ScenarioMessage

	loop := true
	var attachedGnb string
	state := ueCtx.MM5G_NULL
	targetState := ueCtx.MM5G_NULL
	for loop {
		select {
		case task := <-taskChan:
			if attachedGnb == "" && task.TaskType != AttachToGNB {
				msg := fmt.Sprintf("UE %d is not attached to a GNB", ueId)
				log.Error(msg)
				reportChan <- report{success: false, reason: msg}
			}
			switch task.TaskType {
			case AttachToGNB:
				if gnbs[task.Parameters.GnbId] != nil {
					// Create a new UE coroutine
					// ue.NewUE returns context of the new UE
					ueTx = ue.NewUE(ueCfg, ueId, ueRx, gnbs[task.Parameters.GnbId].GetInboundChannel(), wg)
					attachedGnb = task.Parameters.GnbId
					reportChan <- report{success: true}
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					reportChan <- report{success: false, reason: msg}
				}
			case Registration:
				time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() { ueRx <- procedures.UeTesterMessage{Type: procedures.Registration} })
				targetState = ueCtx.MM5G_REGISTERED
			case Deregistration:
				time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() { ueRx <- procedures.UeTesterMessage{Type: procedures.Deregistration} })
				targetState = ueCtx.MM5G_DEREGISTERED
			case Idle:
				time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() { ueRx <- procedures.UeTesterMessage{Type: procedures.Idle} })
				targetState = ueCtx.MM5G_IDLE
			case NewPDUSession:
				if state == ueCtx.MM5G_REGISTERED { // TODO: check if it's possible to send pdu req without being registrated (for negative tests)
					time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() { ueRx <- procedures.UeTesterMessage{Type: procedures.NewPDUSession} })
					if !task.Hang {
						reportChan <- report{success: true, reason: "Need to implement way to check if pduSession is successful"}
					}
				} else {
					msg := fmt.Sprintf("UE %d is not registered", ueId)
					log.Error(msg)
					reportChan <- report{success: false, reason: msg}
				}
			case XNHandover:
				if gnbs[task.Parameters.GnbId] != nil {
					time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
						trigger.TriggerXnHandover(gnbs[attachedGnb], gnbs[task.Parameters.GnbId], int64(ueId))
					})
					targetState = ueCtx.MM5G_REGISTERED
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					reportChan <- report{success: false, reason: msg}
				}
			case NGAPHandover:
				if gnbs[task.Parameters.GnbId] != nil {
					time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
						trigger.TriggerNgapHandover(gnbs[attachedGnb], gnbs[task.Parameters.GnbId], int64(ueId))
					})
					targetState = ueCtx.MM5G_REGISTERED
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					reportChan <- report{success: false, reason: msg}
				}
			case Terminate:
				time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
					ueRx <- procedures.UeTesterMessage{Type: procedures.
						Terminate}
					ueRx = nil
				})
			case Kill:
				time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
					ueRx <- procedures.UeTesterMessage{Type: procedures.
						Kill}
					ueRx = nil
					state = ueCtx.MM5G_NULL
					attachedGnb = ""
				})
			}

		case ueMsg := <-ueTx:
			msg := fmt.Sprint("[UE] Switched from state ", state, " to state ", ueMsg.StateChange)
			log.Info(msg)
			if ueMsg.StateChange == targetState {
				reportChan <- report{success: true, reason: msg}
			}
			if ueMsg.StateChange == ueCtx.MM5G_NULL {
				loop = false
			}
			state = ueMsg.StateChange
		}
	}
}
