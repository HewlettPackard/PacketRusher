/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelEnabled bool, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeHandover int, numPduSessions int) {
	if tunnelEnabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
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
	if numGnb <= 1 && timeBeforeHandover != 0 {
		log.Warn("[TESTER] We are increasing the number of gNodeB to two for handover test cases. Make you sure you fill the requirements for having two gNodeBs.")
		numGnb++
	}
	gnbs := tools.CreateGnbs(numGnb, cfg, &wg)

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(1 * time.Second)

	cfg.Ue.TunnelEnabled = tunnelEnabled

	scenarioChans := make([]chan procedures.UeTesterMessage, numUes+1)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	ueSimCfg := tools.UESimulationConfig{
		Gnbs:                     gnbs,
		Cfg:                      cfg,
		TimeBeforeDeregistration: timeBeforeDeregistration,
		TimeBeforeHandover:       timeBeforeHandover,
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
