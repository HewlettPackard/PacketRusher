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

type ScenarioManager struct {
	gnbs map[string]*context.GNBContext
	ues  map[int]*ueScenarioCtx
}

// Description of the scenario to be run for one UE
type UEScenario struct {
	Config    config.Ue
	Tasks     []Task
	Loop      int // Time between two loops
	ForceStop int // Time before forcefully stoping scenario (0 to desactivate)
}

// Context of running ue scenario
type ueScenarioCtx struct {
	ongoingTask bool
	nextTasks   []Task
	taskChan    chan Task
	reportChan  chan report
	loopTicker  *time.Ticker
	endTicker   *time.Ticker
}

type report struct {
	success bool
	reason  string
}

func (s *ScenarioManager) Start(gnbConfs []config.GNodeB, amfConf config.AMF, ueScenarios []UEScenario, maxRequestRate int) {

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

	s.ues = make(map[int]*ueScenarioCtx)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)
	stopSignal := false
	for !stopSignal {
		done := true
		for ueId := 1; !stopSignal && ueId <= len(ueScenarios); ueId++ {
			if s.ues[ueId] == nil {
				ue := ueScenarioCtx{
					ongoingTask: false,
					nextTasks:   ueScenarios[ueId-1].Tasks,
					taskChan:    make(chan Task, 1),
					reportChan:  make(chan report, 1),
				}

				if ueScenarios[ueId-1].Loop > 0 {
					ue.loopTicker = time.NewTicker(time.Duration(ueScenarios[ueId-1].Loop) * time.Millisecond)
				}
				if ueScenarios[ueId-1].ForceStop > 0 {
					ue.endTicker = time.NewTicker(time.Duration(ueScenarios[ueId-1].ForceStop) * time.Millisecond)
				}

				executer := ueTaskExecuter{
					UeId:       ueId,
					UeCfg:      ueScenarios[ueId-1].Config,
					TaskChan:   ue.taskChan,
					ReportChan: ue.reportChan,
					Gnbs:       s.gnbs,
					Wg:         &wg,
				}
				executer.Run()

				s.ues[ueId] = &ue
			}

			select {
			case report := <-s.ues[ueId].reportChan:

				if !report.success {
					s.ues[ueId].nextTasks = []Task{}
					log.Debug("------------------------------------------- Got report: " + report.reason + " ------------------------------------------- ")
				}
				s.ues[ueId].ongoingTask = false
			case <-sigStop:
				stopSignal = true
			default:
			}
			if ueScenarios[ueId-1].ForceStop > 0 {
				select {
				case <-s.ues[ueId].endTicker.C:
					done = true
					log.Debugf("[UE-%d]Force end of scenario", ueId)
					continue
				default:
				}
			}
			if ueScenarios[ueId-1].Loop > 0 {
				done = false
				select {
				case <-s.ues[ueId].loopTicker.C:
					log.Debugf("[UE-%d] starting new loop", ueId)
					s.ues[ueId].taskChan <- Task{TaskType: Kill}
					s.ues[ueId].nextTasks = ueScenarios[ueId-1].Tasks
					s.ues[ueId].ongoingTask = true
					continue
				default:
				}
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

	time.Sleep(time.Second * 3)
}

type ueTaskExecuter struct {
	UeId       int
	UeCfg      config.Ue
	TaskChan   chan Task
	ReportChan chan report
	Gnbs       map[string]*context.GNBContext
	Wg         *sync.WaitGroup

	running     bool
	ueRx        chan procedures.UeTesterMessage
	ueTx        chan scenario.ScenarioMessage
	attachedGnb string
	timedTasks  map[Task]*time.Timer
	state       int
	targetState int
	lock        sync.Mutex
	toBeKilled  chan procedures.UeTesterMessage
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
	if e.Wg == nil {
		log.Errorf("simulation for ue %d failed: no WaitGroup is given", e.UeId)
		return
	}

	e.running = true
	e.timedTasks = map[Task]*time.Timer{}
	e.state = ueCtx.MM5G_NULL
	e.targetState = ueCtx.MM5G_NULL
	go e.listen()
}

func (e *ueTaskExecuter) listen() {
	log.Infof("simulation started for ue %d", e.UeId)
	loop := true
	for loop {
		select {
		case task := <-e.TaskChan:
			log.Debugf("------------------------------------------- received task: %s -------------------------------------------  ", task.TaskType.ToStr())
			if e.attachedGnb == "" && task.TaskType != AttachToGNB {
				msg := fmt.Sprintf("UE %d is not attached to a GNB", e.UeId)
				log.Error(msg)
				e.ReportChan <- report{success: false, reason: msg}
			}
			switch task.TaskType {
			case AttachToGNB:
				if e.Gnbs[task.Parameters.GnbId] != nil {
					// Create a new UE coroutine
					// ue.NewUE returns context of the new UE
					e.lock.Lock()
					e.ueRx = make(chan procedures.UeTesterMessage)
					e.ueTx = ue.NewUE(e.UeCfg, e.UeId, e.ueRx, e.Gnbs[task.Parameters.GnbId].GetInboundChannel(), e.Wg)
					e.Wg.Add(1)
					e.attachedGnb = task.Parameters.GnbId
					e.lock.Unlock()
					e.ReportChan <- report{success: true}
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					e.ReportChan <- report{success: false, reason: msg}
				}
			case Registration:
				if e.toBeKilled != nil {
					e.toBeKilled <- procedures.UeTesterMessage{Type: procedures.Kill}
				}
				e.sendProcedureToUE(task, procedures.Registration)
				e.targetState = ueCtx.MM5G_REGISTERED
			case Deregistration:
				e.sendProcedureToUE(task, procedures.Deregistration)
				e.targetState = ueCtx.MM5G_DEREGISTERED
			case Idle:
				e.sendProcedureToUE(task, procedures.Idle)
				e.targetState = ueCtx.MM5G_IDLE
			case NewPDUSession:
				if e.state == ueCtx.MM5G_REGISTERED {
					e.sendProcedureToUE(task, procedures.NewPDUSession)
					if !task.Hang {
						// TODO: Implement way to check if pduSession is successful
						e.ReportChan <- report{success: true, reason: "sent PDU session request"}
					}
				} else {
					msg := fmt.Sprintf("UE %d is not registered", e.UeId)
					log.Error(msg)
					e.ReportChan <- report{success: false, reason: msg}
				}
			case XNHandover:
				if e.Gnbs[task.Parameters.GnbId] != nil {
					time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
						trigger.TriggerXnHandover(e.Gnbs[e.attachedGnb], e.Gnbs[task.Parameters.GnbId], int64(e.UeId))
						e.attachedGnb = task.Parameters.GnbId
						// Wait for succesful handover
						// TODO: We should wait for confimation from gnb instead of timer
						time.Sleep(2 * time.Second)
						e.ReportChan <- report{success: true, reason: "succesful XNHandover"}
					})
					e.targetState = ueCtx.MM5G_REGISTERED
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					e.ReportChan <- report{success: false, reason: msg}
				}
			case NGAPHandover:
				if e.Gnbs[task.Parameters.GnbId] != nil {
					time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
						trigger.TriggerNgapHandover(e.Gnbs[e.attachedGnb], e.Gnbs[task.Parameters.GnbId], int64(e.UeId))
						e.attachedGnb = task.Parameters.GnbId
						// Wait for succesful handover
						// TODO: We should wait for confimation from gnb instead of timer
						time.Sleep(2 * time.Second)
						e.ReportChan <- report{success: true, reason: "succesful NGAPHandover"}
					})
					e.targetState = ueCtx.MM5G_REGISTERED
				} else {
					msg := fmt.Sprintf("GNB %s not found", task.Parameters.GnbId)
					log.Error(msg)
					e.ReportChan <- report{success: false, reason: msg}
				}
			case Terminate:
				e.timedTasks[task] = time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
					e.lock.Lock()
					defer e.lock.Unlock()
					if e.toBeKilled != nil {
						e.toBeKilled <- procedures.UeTesterMessage{Type: procedures.Terminate}
					}
					e.ueRx <- procedures.UeTesterMessage{Type: procedures.Terminate}
					e.ueRx = nil
					e.ueTx = nil
					e.state = ueCtx.MM5G_NULL
					e.attachedGnb = ""
					for t := range e.timedTasks {
						e.timedTasks[t].Stop()
						delete(e.timedTasks, t)
					}
					e.ReportChan <- report{success: true, reason: "UE Terminated"}
				})
			case Kill:
				e.timedTasks[task] = time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
					e.lock.Lock()
					defer e.lock.Unlock()
					e.toBeKilled = e.ueRx
					e.ueRx = nil
					e.ueTx = nil
					e.state = ueCtx.MM5G_NULL
					e.attachedGnb = ""
					for t := range e.timedTasks {
						e.timedTasks[t].Stop()
						delete(e.timedTasks, t)
					}
					e.ReportChan <- report{success: true, reason: "UE Killed"}
				})
			}

		case ueMsg, open := <-e.ueTx:
			if open {
				msg := fmt.Sprintf("[UE-%d] Switched from state %d to state %d ", e.UeId, e.state, ueMsg.StateChange)
				log.Info(msg)
				if ueMsg.StateChange == e.targetState {
					e.ReportChan <- report{success: true, reason: msg}
				}
				e.state = ueMsg.StateChange
			}
		}
	}
	e.running = false
}

func (e *ueTaskExecuter) sendProcedureToUE(task Task, procedure procedures.UeTesterMessageType) {
	if task.Delay > 0 {
		e.timedTasks[task] = time.AfterFunc(time.Duration(task.Delay)*time.Millisecond, func() {
			e.ueRx <- procedures.UeTesterMessage{Type: procedure}
			e.lock.Lock()
			e.timedTasks[task].Stop()
			delete(e.timedTasks, task)
			e.lock.Unlock()
		})
	} else {
		e.ueRx <- procedures.UeTesterMessage{Type: procedure}
	}
}
