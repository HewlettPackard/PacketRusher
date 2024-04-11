/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"errors"
	"fmt"
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

var (
	InitGnb = gnb.InitGnb
)

type ScenarioManager struct {
	gnbs              map[string]*context.GNBContext
	ues               map[int]*ueTasksCtx
	reportChan        chan report
	registrationQueue chan order
	wg                *sync.WaitGroup
	sigStop           chan os.Signal
	ueScenarios       []UEScenario
	stop              bool
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

func Start(gnbConfs []config.GNodeB, amfConf config.AMF, ueScenarios []UEScenario, timeBetweenRegistration int, maxRequestRate int) {
	log.Info("------------------------------------ Starting scenario ------------------------------------")
	s := initManager(ueScenarios, timeBetweenRegistration)
	sender.Init(maxRequestRate)
	s.setupGnbs(gnbConfs, amfConf)
	s.setupScenarios(ueScenarios)
	s.executeScenarios()
	s.cleanup()
	log.Info("------------------------------------ Scenario finished ------------------------------------")
}

func initManager(ueScenarios []UEScenario, timeBetweenRegistration int) *ScenarioManager {
	s := &ScenarioManager{}
	s.wg = &sync.WaitGroup{}
	s.registrationQueue = make(chan order, len(ueScenarios))
	go startOrderRateController(s.registrationQueue, timeBetweenRegistration)
	s.reportChan = make(chan report, 10)
	s.ues = make(map[int]*ueTasksCtx)
	s.sigStop = make(chan os.Signal, 1)
	signal.Notify(s.sigStop, os.Interrupt)
	return s
}

func (s *ScenarioManager) setupScenarios(ueScenarios []UEScenario) {
	s.stop = false
	s.ueScenarios = ueScenarios
	for ueId := 1; ueId <= len(ueScenarios); ueId++ {
		err := s.setupUeTaskExecuter(ueId, ueScenarios[ueId-1])
		log.Errorf("scenario for UE %d could not be started: %v", ueId, err)
		select {
		case <-s.sigStop:
			s.stop = true
			ueId = len(ueScenarios) + 1 // exit loop
		default:
		}
	}
}

func (s *ScenarioManager) executeScenarios() {
	for !s.stop {
		s.stop = true
		select {
		case <-s.sigStop:
			s.stop = true
		case report := <-s.reportChan:
			s.handleReport(report)
			if err := s.manageNextTask(report.ueId, s.ueScenarios[report.ueId-1]); err != nil {
				for i := range s.ues {
					if !s.ues[i].done {
						s.stop = false
					}
				}
			}
		}
	}
}

func (s *ScenarioManager) cleanup() {
	for _, ue := range s.ues {
		ue.workerChan <- Task{
			TaskType: Terminate,
		}
	}
	close(s.reportChan)
	close(s.registrationQueue)
	time.Sleep(time.Second * 3)
}

func (s *ScenarioManager) handleReport(report report) {
	log.Debugf("------------------------------------ Report UE-%d ------------------------------------", report.ueId)
	if !report.success {
		s.ues[report.ueId].nextTasks = []Task{}
		log.Errorf("[UE-%d] Reported error:  %s", report.ueId, report.reason)
		return
	}
	log.Infof("[UE-%d] Reported succes:  %s", report.ueId, report.reason)
}

func (s *ScenarioManager) manageNextTask(ueId int, ueScenarios UEScenario) error {
	if s.sendNextTask(ueId) == nil {
		return nil
	}
	if ueScenarios.Loop {
		s.restartUeScenario(ueId, ueScenarios)
		return nil
	}
	return errors.New("no task to be managed")
}

func (s *ScenarioManager) sendNextTask(ueId int) error {
	_, exist := s.ues[ueId]
	if !exist {
		return errors.New("ue not found")
	}
	if len(s.ues[ueId].nextTasks) > 0 {
		if s.ues[ueId].nextTasks[0].TaskType == Registration {
			s.registrationQueue <- order{task: s.ues[ueId].nextTasks[0], workerChan: s.ues[ueId].workerChan}
		} else {
			s.ues[ueId].workerChan <- s.ues[ueId].nextTasks[0]
		}
		s.ues[ueId].nextTasks = s.ues[ueId].nextTasks[1:]
		return nil
	}
	return errors.New("ue does not have task to do")
}

func (s *ScenarioManager) restartUeScenario(ueId int, ueScenario UEScenario) error {
	_, exist := s.ues[ueId]
	if !exist {
		return errors.New("ue not found")
	}
	if ueScenario.Tasks == nil {
		return fmt.Errorf("ue %d tasks is nil", ueId)
	}
	log.Debugf("[UE-%d] starting new loop", ueId)
	s.ues[ueId].workerChan <- Task{TaskType: Kill}
	s.ues[ueId].nextTasks = ueScenario.Tasks
	return nil
}

func (s *ScenarioManager) setupGnbs(gnbConfs []config.GNodeB, amfConf config.AMF) {
	s.gnbs = make(map[string]*context.GNBContext)

	for gnbConf := range gnbConfs {
		s.gnbs[gnbConfs[gnbConf].PlmnList.GnbId] = InitGnb(gnbConfs[gnbConf], amfConf, s.wg)
		s.wg.Add(1)
	}

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(2 * time.Second)
}

func (s *ScenarioManager) setupUeTaskExecuter(ueId int, ueScenario UEScenario) error {
	if ueId < 1 {
		return errors.New("ueId must be at least 1")
	}
	_, exist := s.ues[ueId]
	if exist {
		return errors.New("this ue already have a task executer")
	}
	if ueScenario.Tasks == nil {
		return errors.New("tasks list is nil")
	}
	if ueScenario.Config == (config.Ue{}) {
		return errors.New("config is empty")
	}
	if s.reportChan == nil {
		return errors.New("scenario manager's report channel is not set")
	}
	if s.gnbs == nil {
		return errors.New("scenario manager's gnb list is not set")
	}
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
	return nil
}

func startOrderRateController(orderChan <-chan order, timeBetweenRegistration int) {
	waitTime := time.Duration(timeBetweenRegistration) * time.Millisecond
	orderTicker := time.NewTicker(waitTime)
	var order order
	var open bool
	for {
		<-orderTicker.C
		orderTicker.Stop()
		order, open = <-orderChan
		if open {
			order.workerChan <- order.task
			orderTicker.Reset(waitTime)
		} else {
			return
		}
	}
}
