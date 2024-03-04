/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/scenario"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeReconnecting int, numPduSessions int, maxRequestRate int) {
	if tunnelMode != config.TunnelDisabled {
		if !dedicatedGnb {
			log.Fatal("You cannot use the --tunnel option, without using the --dedicatedGnb option")
		}
		if timeBetweenRegistration < 500 {
			log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
		}
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	sender.Init(maxRequestRate)

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	var numGnb int
	if dedicatedGnb {
		numGnb = numUes
	} else {
		numGnb = 1
	}
	if numGnb <= 1 && (timeBeforeXnHandover != 0 || timeBeforeNgapHandover != 0) {
		log.Warn("[TESTER] We are increasing the number of gNodeB to two for handover test cases. Make you sure you fill the requirements for having two gNodeBs.")
		numGnb++
	}
	gnbs := tools.CreateGnbs(numGnb, cfg, &wg)

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(5 * time.Second)

	cfg.Ue.TunnelMode = tunnelMode

	scenarioChans := make([]chan procedures.UeTesterMessage, numUes+1)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	ueSimCfg := tools.UESimulationConfig{
		Gnbs:                     gnbs,
		Cfg:                      cfg,
		TimeBeforeDeregistration: timeBeforeDeregistration,
		TimeBeforeNgapHandover:   timeBeforeNgapHandover,
		TimeBeforeXnHandover:     timeBeforeXnHandover,
		TimeBeforeIdle:           timeBeforeIdle,
		TimeBeforeReconnecting:   timeBeforeReconnecting,
		NumPduSessions:           numPduSessions,
	}

	stopSignal := true
	for stopSignal {
		// If CTRL-C signal has been received,
		// stop creating new UEs, else we create numUes UEs
		for ueSimCfg.UeId = 1; stopSignal && ueSimCfg.UeId <= numUes; ueSimCfg.UeId++ {
			// If there is currently a coroutine handling current UE
			// kill it, before creating a new coroutine with same UE
			// Use case: Registration of N UEs in loop, when loop = true
			if scenarioChans[ueSimCfg.UeId] != nil {
				scenarioChans[ueSimCfg.UeId] <- procedures.UeTesterMessage{Type: procedures.Kill}
				close(scenarioChans[ueSimCfg.UeId])
				scenarioChans[ueSimCfg.UeId] = nil
			}
			scenarioChans[ueSimCfg.UeId] = make(chan procedures.UeTesterMessage)
			ueSimCfg.ScenarioChan = scenarioChans[ueSimCfg.UeId]

			tools.SimulateSingleUE(ueSimCfg, &wg)

			// Before creating a new UE, we wait for timeBetweenRegistration ms
			time.Sleep(time.Duration(timeBetweenRegistration) * time.Millisecond)

			select {
			case <-sigStop:
				stopSignal = false
			default:
			}
		}
		// If loop = false, we don't go over the for loop a second time
		// and we only do the numUes registration once
		if !loop {
			break
		}
	}

	if stopSignal {
		<-sigStop
	}
	for _, scenarioChan := range scenarioChans {
		if scenarioChan != nil {
			scenarioChan <- procedures.UeTesterMessage{Type: procedures.Terminate}
		}
	}

	time.Sleep(time.Second * 1)
}

func TestSingleUe(tunnelMode config.TunnelMode, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, numPduSessions int) {
	if tunnelMode != config.TunnelDisabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cfg := config.GetConfig()

	tasks := []scenario.Task{
		{
			TaskType: scenario.AttachToGNB,
			Parameters: struct {
				GnbId string
			}{cfg.GNodeB.PlmnList.GnbId},
		},
		{
			TaskType: scenario.Registration,
		},
	}

	for i := 0; i < numPduSessions; i++ {
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.NewPDUSession,
		})
	}

	if timeBeforeNgapHandover != 0 {
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.NGAPHandover,
			Delay:    timeBeforeNgapHandover,
		})
	}

	if timeBeforeXnHandover != 0 {
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.XNHandover,
			Delay:    timeBeforeXnHandover,
		})
	}
	if timeBeforeIdle != 0 {
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.Idle,
			Delay:    timeBeforeIdle,
		})
	}

	if timeBeforeDeregistration != 0 {
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.Deregistration,
			Delay:    timeBeforeDeregistration,
		})
	}

	gnbs := []config.GNodeB{cfg.GNodeB}
	ueScenario := scenario.UEScenario{
		Config: cfg.Ue,
		Tasks:  tasks,
	}
	ueScenarios := []scenario.UEScenario{ueScenario}

	r := scenario.Runner{}
	r.Start(gnbs, cfg.AMF, ueScenarios, 0)
}
