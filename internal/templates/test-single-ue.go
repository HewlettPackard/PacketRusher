/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/scenario"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func TestSingleUe(tunnelMode config.TunnelMode, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, numPduSessions int) {
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
		
	// TODO: use arg position to determine order of procedures
	if timeBeforeNgapHandover != 0 {
		nextgnb = (nextgnb + 1) % 2
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.NGAPHandover,
			Delay:    timeBeforeNgapHandover,
			Parameters: struct {
				GnbId string
			}{gnbs[nextgnb].PlmnList.GnbId},
		})
	}

	if timeBeforeXnHandover != 0 {
		nextgnb = (nextgnb + 1) % 2
		tasks = append(tasks, scenario.Task{
			TaskType: scenario.XNHandover,
			Delay:    timeBeforeXnHandover,
			Parameters: struct {
				GnbId string
			}{gnbs[nextgnb].PlmnList.GnbId},
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
