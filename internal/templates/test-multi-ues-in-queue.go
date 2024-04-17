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
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeServiceRequest int, numPduSessions int) {
	if tunnelMode != config.TunnelDisabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cfg := config.GetConfig()
	cfg.Ue.TunnelMode = tunnelMode

	gnb := cfg.GNodeB
	gnbs := []config.GNodeB{gnb}
	nextgnb := 0
	if timeBeforeNgapHandover > 0 || timeBeforeXnHandover > 0 {
		gnb = nextGnbConf(gnb, 1, cfg.GNodeB.PlmnList.GnbId)
		gnbs = append(gnbs, gnb)
	}
	ueScenarios := []scenario.UEScenario{}
	for i := 0; i < numUes; i++ {
		if dedicatedGnb && i != 0 {
			gnb = nextGnbConf(gnb, len(gnbs), cfg.GNodeB.PlmnList.GnbId)
			gnbs = append(gnbs, gnb)
			nextgnb = len(gnbs) - 1
			if timeBeforeNgapHandover > 0 || timeBeforeXnHandover > 0 {
				gnb = nextGnbConf(gnb, len(gnbs), cfg.GNodeB.PlmnList.GnbId)
				gnbs = append(gnbs, gnb)
			}
		}

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
			if timeBeforeServiceRequest != 0 {
				tasks = append(tasks, scenario.Task{
					TaskType: scenario.ServiceRequest,
					Delay:    timeBeforeIdle + timeBeforeServiceRequest,
				})
			}
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
		for j := 0; j < len(tasks); j++ {
			tasks[j].Delay = tasks[j].Delay - sumDelay
			sumDelay += tasks[j].Delay

			if tasks[j].TaskType == scenario.NGAPHandover || tasks[j].TaskType == scenario.XNHandover {
				tasks[j].Parameters.GnbId = gnbs[(len(gnbs)-2)+(nextgnb+1)%2].PlmnList.GnbId
			}
		}

		ueCfg := cfg.Ue
		ueCfg.Msin = tools.IncrementMsin(i+1, cfg.Ue.Msin)
		ueScenario := scenario.UEScenario{
			Config: ueCfg,
			Tasks:  tasks,
		}

		ueScenario.Loop = loop
		if tasks[len(tasks)-1].TaskType != scenario.Deregistration {
			ueScenario.Hang = true
		}

		ueScenarios = append(ueScenarios, ueScenario)
	}

	scenario.Start(gnbs, cfg.AMFs, ueScenarios, timeBetweenRegistration, 0)
}

func nextGnbConf(gnb config.GNodeB, i int, baseId string) config.GNodeB {
	var err error
	gnb.PlmnList.GnbId = genereateGnbId(i, baseId)
	gnb.ControlIF.Ip, err = tools.IncrementIP(gnb.ControlIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N2: " + err.Error())
	}
	gnb.DataIF.Ip, err = tools.IncrementIP(gnb.DataIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N3: " + err.Error())
	}
	return gnb
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
