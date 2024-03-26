/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ScenarioManager struct {
	gnbs              map[string]*context.GNBContext
	ues               map[int]*ueTasksCtx
	reportChan        chan report
	registrationQueue chan order
	wg                *sync.WaitGroup // TODO: not used, to remove
}

// Context of running ue scenario
type ueTasksCtx struct {
	done       bool
	nextTasks  []Task
	workerChan chan Task
}

type report struct {
	ueId    int
	success bool
	reason  string
}

type order struct {
	task       Task
	workerChan chan Task
}

func (s *ScenarioManager) StartScenario(gnbConfs []config.GNodeB, amfConf config.AMF, ueScenarios []UEScenario, timeBetweenRegistration int, maxRequestRate int) {
	log.Info("------------------------------------ Starting scenario ------------------------------------")
	s.wg = &sync.WaitGroup{}

	s.reportChan = make(chan report, 10)
	s.registrationQueue = make(chan order, len(ueScenarios))
	stopChan := make(chan int, 1)
	go startRegistrationRateController(s.registrationQueue, timeBetweenRegistration, stopChan)

	sender.Init(maxRequestRate)
	s.setupGnbs(gnbConfs, amfConf)
	s.ues = make(map[int]*ueTasksCtx)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	endSenario := false
	for ueId := 1; ueId <= len(ueScenarios); ueId++ {
		s.setupUeTaskExecuter(ueId, ueScenarios[ueId-1])

		select {
		case <-sigStop:
			endSenario = true
			ueId = len(ueScenarios) + 1 // exit loop
		default:
		}
	}

	for !endSenario {
		endSenario = true
		select {
		case <-sigStop:
			endSenario = true
		case report := <-s.reportChan:
			s.handleReport(report)
			s.manageNextTask(report.ueId, ueScenarios)
			for i := range s.ues {
				if !s.ues[i].done {
					endSenario = false
				}
			}
		}
	}

	for _, ue := range s.ues {
		ue.workerChan <- Task{
			TaskType: Terminate,
		}
	}
	stopChan <- 0

	time.Sleep(time.Second * 3)
	log.Info("------------------------------------ Scenario finished ------------------------------------")
}

func (s *ScenarioManager) handleReport(report report) {
	log.Debugf("------------------------------------ Report UE-%d ------------------------------------", report.ueId)
	if !report.success {
		s.ues[report.ueId].nextTasks = []Task{}
		log.Errorf("[UE-%d] Reported error:  %s", report.ueId, report.reason)
	}
	log.Infof("[UE-%d] Reported succes:  %s", report.ueId, report.reason)
}

func (s *ScenarioManager) manageNextTask(ueId int, ueScenarios []UEScenario) bool {
	newTask := false
	if len(s.ues[ueId].nextTasks) > 0 {
		s.sendNextTask(ueId)
		return true
	}
	if ueScenarios[ueId-1].Loop {
		s.restartUeScenario(ueId, ueScenarios[ueId-1])
		return true
	}
	s.ues[ueId].done = true
	return newTask
}

func (s *ScenarioManager) sendNextTask(ueId int) {
	if s.ues[ueId].nextTasks[0].TaskType == Registration {
		s.registrationQueue <- order{task: s.ues[ueId].nextTasks[0], workerChan: s.ues[ueId].workerChan}
	} else {
		s.ues[ueId].workerChan <- s.ues[ueId].nextTasks[0]
	}
	s.ues[ueId].nextTasks = s.ues[ueId].nextTasks[1:]
}

func (s *ScenarioManager) restartUeScenario(ueId int, ueScenario UEScenario) {
	log.Debugf("[UE-%d] starting new loop", ueId)
	s.ues[ueId].workerChan <- Task{TaskType: Kill}
	s.ues[ueId].nextTasks = ueScenario.Tasks
}

func (s *ScenarioManager) setupGnbs(gnbConfs []config.GNodeB, amfConf config.AMF) {
	s.gnbs = make(map[string]*context.GNBContext)

	for gnbConf := range gnbConfs {
		s.gnbs[gnbConfs[gnbConf].PlmnList.GnbId] = gnb.InitGnb2(gnbConfs[gnbConf], amfConf, s.wg) // TODO: replace InitGNB2 with InitGNB
		s.wg.Add(1)
	}

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(2 * time.Second)
}

func (s *ScenarioManager) setupUeTaskExecuter(ueId int, ueScenario UEScenario) {
	ue := ueTasksCtx{
		done:       false,
		nextTasks:  ueScenario.Tasks,
		workerChan: make(chan Task, 1),
	}

	if ueScenario.ForceStop > 0 {
		time.AfterFunc(time.Duration(ueScenario.ForceStop)*time.Millisecond, func() {
			ue.workerChan <- Task{TaskType: Terminate}
			ue.done = true
		})
	}

	worker := ueTaskExecuter{
		UeId:       ueId,
		UeCfg:      ueScenario.Config,
		TaskChan:   ue.workerChan,
		ReportChan: s.reportChan,
		Gnbs:       s.gnbs,
		wg:         s.wg,
	}
	worker.Run()

	s.ues[ueId] = &ue
}

func startRegistrationRateController(orderChan chan order, timeBetweenRegistration int, stopChan chan int) {
	waitTime := time.Duration(timeBetweenRegistration) * time.Millisecond
	registrationTicker := time.NewTicker(waitTime)
	log.Debug("Started registration rate controller")
	for {
		log.Debug("starting new register ticker loop")
		select {
		case <-registrationTicker.C:
			log.Debug("registration rate controller tick! stoping ticker")
			registrationTicker.Stop()
			select {
			case order := <-orderChan:
				log.Debug("found register order, forwading task and resetting timer")
				order.workerChan <- order.task
				registrationTicker.Reset(waitTime)
			case <-stopChan:
				return
			}
		case <-stopChan:
			return
		}
	}
}
