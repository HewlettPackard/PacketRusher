/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	ueCtx "my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ueTaskExecuter struct {
	UeId       int
	UeCfg      config.Ue
	TaskChan   chan Task
	ReportChan chan report
	Gnbs       map[string]*context.GNBContext

	wg                   *sync.WaitGroup
	running              bool
	ueRx                 chan procedures.UeTesterMessage
	ueTx                 chan scenario.ScenarioMessage
	attachedGnb          string
	timedTasks           map[Task]*time.Timer
	lastReceivedTaskTime time.Time
	state                int
	targetState          int
	lock                 sync.Mutex
	toBeKilled           chan procedures.UeTesterMessage
	loop                 bool
}

func (e *ueTaskExecuter) Run() {
	if e.running {
		log.Errorf("simulation for ue %d failed: already running", e.UeId)
		return
	}
	if e.UeCfg == (config.Ue{}) {
		log.Errorf("simulation for ue %d failed: no config", e.UeId)
		return
	}
	if e.TaskChan == nil {
		log.Errorf("simulation for ue %d failed: no taskChan", e.UeId)
		return
	}
	if e.ReportChan == nil {
		log.Errorf("simulation for ue %d failed: no reportChan", e.UeId)
		return
	}
	if e.Gnbs == nil {
		log.Errorf("simulation for ue %d failed: no gnbs", e.UeId)
		return
	}
	if e.wg == nil {
		log.Errorf("simulation for ue %d failed: no WaitGroup is given", e.UeId)
		return
	}

	e.running = true

	e.timedTasks = map[Task]*time.Timer{}
	e.state = ueCtx.MM5G_NULL
	e.targetState = ueCtx.MM5G_NULL
	e.lastReceivedTaskTime = time.Now()
	go e.listen()
}

func (e *ueTaskExecuter) listen() {
	e.loop = true
	log.Infof("simulation started for ue %d", e.UeId)
	e.sendReport(true, "simulation started")
	for e.loop {
		select {
		case task := <-e.TaskChan:
			e.handleTaskTimer(task)
		case ueMsg, open := <-e.ueTx:
			if open {
				e.handleUeMsg(ueMsg)
			}
		}
	}
	e.running = false
}

func (e *ueTaskExecuter) handleUeMsg(ueMsg scenario.ScenarioMessage) {
	msg := fmt.Sprintf("Switched from state %d to state %d ", e.state, ueMsg.StateChange)
	if ueMsg.StateChange == e.targetState {
		e.sendReport(true, msg)
	}
	e.state = ueMsg.StateChange
}

func (e *ueTaskExecuter) handleTaskTimer(task Task) {
	if task.Delay > 0 {
		task.Delay = max(0, task.Delay-int(time.Since(e.lastReceivedTaskTime).Milliseconds()))
		e.timedTasks[task] = time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
			e.handleTask(task)
			e.lock.Lock()
			delete(e.timedTasks, task)
			e.lock.Unlock()
		})
	} else {
		e.handleTask(task)
	}
	e.lastReceivedTaskTime = time.Now()
}

func (e *ueTaskExecuter) handleTask(task Task) {
	if log.GetLevel() >= 5 {
		log.Debugf("------------------------------------ Task UE-%d ------------------------------------", e.UeId)
		log.Debugf("[UE-%d] Received Task: %s", e.UeId, task.TaskType.ToStr())
		log.Debugf("UE %d is attached to %s, tasked to %s", e.UeId, e.attachedGnb, task.TaskType.ToStr())
		for i := range e.Gnbs {
			for j := 1; j <= 3; j++ {
				_, err := e.Gnbs[i].GetGnbUeByPrUeId(int64(j))
				if err == nil {
					log.Debugf("GNB %s has ue %d context", e.Gnbs[i].GetGnbId(), j)
				}
			}
		}
	}

	if e.attachedGnb == "" && task.TaskType != AttachToGNB {
		msg := fmt.Sprintf("UE %d is not attached to a GNB", e.UeId)
		e.sendReport(false, msg)
	}

	e.dispatchTask(task)
}

func (e *ueTaskExecuter) dispatchTask(task Task) {
	switch task.TaskType {
	case AttachToGNB:
		e.handleAttachToGnb(task)
	case Registration:
		e.handleRegistration()
	case Deregistration:
		e.handleDeregistration()
	case Idle:
		e.handleIdle()
	case ServiceRequest:
		e.handleServiceRequest()
	case NewPDUSession:
		e.handleNewPduSession(task)
	case XNHandover:
		e.handleXnHandover(task)
	case NGAPHandover:
		e.handleNgapHandover(task)
	case Terminate:
		e.handleTerminate()
	case Kill:
		e.handleKill()
	default:
		e.sendReport(false, "unknown task")
	}
}

func (e *ueTaskExecuter) handleAttachToGnb(task Task) {
	if e.Gnbs[task.Parameters.GnbId] != nil {
		// Create a new UE coroutine
		// ue.NewUE returns context of the new UE
		e.lock.Lock()
		e.ueRx = make(chan procedures.UeTesterMessage)
		e.ueTx = ue.NewUE(e.UeCfg, e.UeId, e.ueRx, e.Gnbs[task.Parameters.GnbId].GetInboundChannel(), e.wg)
		e.wg.Add(1)
		e.attachedGnb = task.Parameters.GnbId
		e.lock.Unlock()
		e.sendReport(true, "UE successfully attached to GNB "+task.Parameters.GnbId)
	} else {
		msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
		e.sendReport(false, msg)
	}
}

func (e *ueTaskExecuter) handleRegistration() {
	if e.toBeKilled != nil {
		e.toBeKilled <- procedures.UeTesterMessage{Type: procedures.Kill}
	}
	e.ueRx <- procedures.UeTesterMessage{Type: procedures.Registration}
	e.targetState = ueCtx.MM5G_REGISTERED
}

func (e *ueTaskExecuter) handleDeregistration() {
	e.ueRx <- procedures.UeTesterMessage{Type: procedures.Deregistration}
	e.targetState = ueCtx.MM5G_DEREGISTERED
}

func (e *ueTaskExecuter) handleIdle() {
	e.ueRx <- procedures.UeTesterMessage{Type: procedures.Idle}
	e.targetState = ueCtx.MM5G_IDLE
	// TODO Should sync with gnb before sending Idle success report
}

func (e *ueTaskExecuter) handleServiceRequest() {
	e.ueRx <- procedures.UeTesterMessage{Type: procedures.ServiceRequest}
	e.targetState = ueCtx.MM5G_REGISTERED
}

func (e *ueTaskExecuter) handleNewPduSession(task Task) {
	if e.state == ueCtx.MM5G_REGISTERED {
		log.Debugf("UE %d Before PDU session creation", e.UeId)
		for i := range e.Gnbs {
			_, err := e.Gnbs[i].GetGnbUeByPrUeId(int64(e.UeId))
			if err == nil {
				log.Debugf("GNB %s has ue %d", e.Gnbs[i].GetGnbId(), e.UeId)
			} else {
				log.Debugf("GNB %s does not have %d: %s", e.Gnbs[i].GetGnbId(), e.UeId, err)
			}
		}
		e.ueRx <- procedures.UeTesterMessage{Type: procedures.NewPDUSession}
		if !task.Hang {
			// TODO: Implement way to check if pduSession is successful
			time.Sleep(2 * time.Second)
			e.sendReport(true, "sent PDU session request")
		}
	} else {
		msg := fmt.Sprintf("UE %d is not registered", e.UeId)
		e.sendReport(false, msg)
	}
}

func (e *ueTaskExecuter) handleXnHandover(task Task) {
	if e.Gnbs[task.Parameters.GnbId] != nil {
		trigger.TriggerXnHandover(e.Gnbs[e.attachedGnb], e.Gnbs[task.Parameters.GnbId], int64(e.UeId))
		e.attachedGnb = task.Parameters.GnbId
		// Wait for succesful handover
		// TODO: We should wait for confimation from gnb instead of timer
		time.Sleep(2 * time.Second)
		e.sendReport(true, "succesful XNHandover")
		e.targetState = ueCtx.MM5G_REGISTERED
	} else {
		msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
		e.sendReport(false, msg)
	}
}

func (e *ueTaskExecuter) handleNgapHandover(task Task) {
	if e.Gnbs[task.Parameters.GnbId] != nil {
		trigger.TriggerNgapHandover(e.Gnbs[e.attachedGnb], e.Gnbs[task.Parameters.GnbId], int64(e.UeId))
		e.attachedGnb = task.Parameters.GnbId
		// Wait for succesful handover
		// TODO: We should wait for confimation from gnb instead of timer
		time.Sleep(2 * time.Second)
		e.sendReport(true, "succesful NGAPHandover")
		e.targetState = ueCtx.MM5G_REGISTERED
	} else {
		msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
		e.sendReport(false, msg)
	}
}

func (e *ueTaskExecuter) handleTerminate() {
	if e.toBeKilled != nil {
		e.toBeKilled <- procedures.UeTesterMessage{Type: procedures.Terminate}
		e.lock.Lock()
		e.toBeKilled = nil
		e.lock.Unlock()
	}
	e.ueRx <- procedures.UeTesterMessage{Type: procedures.Terminate}
	e.reset()
	e.sendReport(true, "UE Terminated")
	close(e.TaskChan)
	e.loop = false
}

func (e *ueTaskExecuter) handleKill() {
	// Not killing immediately to be able to deregister ue when receiving terminate command right after killing command
	e.lock.Lock()
	e.toBeKilled = e.ueRx
	e.lock.Unlock()
	e.reset()
	e.sendReport(true, "UE Killed")
}

func (e *ueTaskExecuter) reset() {
	e.lock.Lock()
	e.ueRx = nil
	e.ueTx = nil
	e.state = ueCtx.MM5G_NULL
	e.attachedGnb = ""
	for t := range e.timedTasks {
		e.timedTasks[t].Stop()
		delete(e.timedTasks, t)
	}
	e.lock.Unlock()
}

func (e *ueTaskExecuter) sendReport(success bool, msg string) {
	e.ReportChan <- report{success: success, reason: msg, ueId: e.UeId}
}
