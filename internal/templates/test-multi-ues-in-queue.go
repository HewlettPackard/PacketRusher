/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/scenario"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeReconnecting int, numPduSessions int) {
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
	time.Sleep(1 * time.Second)

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

func TestsingleUePdu(tunnelMode config.TunnelMode, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, numPduSessions int) {
	if tunnelMode != config.TunnelDisabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cfg := config.GetConfig()

	var err error
	gnb2 := cfg.GNodeB
	gnb2.PlmnList.GnbId = genereateGnbId(1, cfg.GNodeB.PlmnList.GnbId)

	gnb2.ControlIF.Ip, err = tools.IncrementIP(cfg.GNodeB.ControlIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N2: " + err.Error())
	}
	gnb2.DataIF.Ip, err = tools.IncrementIP(cfg.GNodeB.DataIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N3: " + err.Error())
	}

	gnbs := []config.GNodeB{cfg.GNodeB, gnb2}
	nextgnb := 0

	tasks := []scenario.Task{
		{
			TaskType: scenario.AttachToGNB,
			Parameters: struct {
				GnbId string
			}{gnbs[nextgnb].PlmnList.GnbId},
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

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Delay < tasks[j].Delay
	})

	sumDelay := 0
	for i := 0; i < len(tasks); i++ {
		tasks[i].Delay = tasks[i].Delay - sumDelay
		sumDelay += tasks[i].Delay

		if tasks[i].TaskType == scenario.NGAPHandover || tasks[i].TaskType == scenario.XNHandover {
			nextgnb = (nextgnb + 1) % 2
			tasks[i].Parameters.GnbId = gnbs[nextgnb].PlmnList.GnbId
		}
	}

	ueScenario := scenario.UEScenario{
		Config: cfg.Ue,
		Tasks:  tasks,
	}
	ueScenarios := []scenario.UEScenario{ueScenario}

	if loop {
		ueScenario.Loop = timeBetweenRegistration
	}

	r := scenario.ScenarioManager{}
	r.Start(gnbs, cfg.AMF, ueScenarios, 0)
}

func genereateGnbId(i int, gnbId string) string {

	gnbId_int, err := strconv.ParseInt(gnbId, 16, 0)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given gnbId is invalid")
	}
	base := int(gnbId_int) + i

	gnbId = fmt.Sprintf("%06x", base)
	return gnbId
}
