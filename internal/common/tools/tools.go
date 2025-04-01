/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	gnbCxt "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	ueCtx "my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"strconv"
	"sync"
	"time"

	"errors"

	log "github.com/sirupsen/logrus"
)

func CreateGnbs(count int, cfg config.Config, wg *sync.WaitGroup) map[string]*gnbCxt.GNBContext {
	gnbs := make(map[string]*gnbCxt.GNBContext)
	// Each gNB have their own IP address on both N2 and N3
	// TODO: Limitation for now, these IPs must be sequential, eg:
	// gnb[0].n2_ip = 192.168.2.10, gnb[0].n3_ip = 192.168.3.10
	// gnb[1].n2_ip = 192.168.2.11, gnb[1].n3_ip = 192.168.3.11
	// ...
	baseGnbId := cfg.GNodeB.PlmnList.GnbId
	for i := 1; i <= count; i++ {
		gnbs[cfg.GNodeB.PlmnList.GnbId] = gnb.InitGnb(cfg, wg)
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the incremented IPs, still belong to the interfaces' subnet
		cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i, baseGnbId)
		cfg.GNodeB.ControlIF = cfg.GNodeB.ControlIF.WithNextAddr()
		cfg.GNodeB.DataIF = cfg.GNodeB.DataIF.WithNextAddr()
	}
	return gnbs
}

func IncrementIP(origIP, cidr string) (string, error) {
	ip := net.ParseIP(origIP)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return origIP, err
	}
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
	if !ipNet.Contains(ip) {
		return origIP, errors.New("Ip is not in provided subnet")
	}
	return ip.String(), nil
}

func gnbIdGenerator(i int, gnbId string) string {

	gnbId_int, err := strconv.ParseInt(gnbId, 16, 0)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given gnbId is invalid")
	}
	base := int(gnbId_int) + i

	gnbId = fmt.Sprintf("%06X", base)
	return gnbId
}

type UESimulationConfig struct {
	UeId                     int
	Gnbs                     map[string]*gnbCxt.GNBContext
	Cfg                      config.Config
	ScenarioChan             chan procedures.UeTesterMessage
	TimeBeforeDeregistration int
	TimeBeforeNgapHandover   int
	TimeBeforeXnHandover     int
	TimeBeforeIdle           int
	TimeBeforeReconnecting   int
	NumPduSessions           int
	RegistrationLoop         bool
	LoopCount                int
	TimeBeforeReregistration int
}

func SimulateSingleUE(simConfig UESimulationConfig, wg *sync.WaitGroup) {
	numGnb := len(simConfig.Gnbs)
	ueCfg := simConfig.Cfg
	ueCfg.Ue.Msin = IncrementMsin(simConfig.UeId, simConfig.Cfg.Ue.Msin)
	log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", ueCfg.Ue.Msin, " UE")

	gnbIdGen := func(index int) string {
		return gnbIdGenerator((simConfig.UeId+index)%numGnb, ueCfg.GNodeB.PlmnList.GnbId)
	}

	// Launch a coroutine to handle UE's individual scenario
	go func(scenarioChan chan procedures.UeTesterMessage, ueId int) {
		i := 0
		for {
			i++
			wg.Add(1)

			ueRx := make(chan procedures.UeTesterMessage)

			// Create a new UE coroutine
			// ue.NewUE returns context of the new UE
			ueTx := ue.NewUE(ueCfg, ueId, ueRx, simConfig.Gnbs[gnbIdGen(0)].GetInboundChannel(), wg)

			// We tell the UE to perform a registration
			ueRx <- procedures.UeTesterMessage{Type: procedures.Registration}

			var deregistrationChannel <-chan time.Time = nil
			if simConfig.TimeBeforeDeregistration != 0 {
				deregistrationChannel = time.After(time.Duration(simConfig.TimeBeforeDeregistration) * time.Millisecond)
			}

			nextHandoverId := 0
			var ngapHandoverChannel <-chan time.Time = nil
			if simConfig.TimeBeforeNgapHandover != 0 {
				ngapHandoverChannel = time.After(time.Duration(simConfig.TimeBeforeNgapHandover) * time.Millisecond)
			}
			var xnHandoverChannel <-chan time.Time = nil
			if simConfig.TimeBeforeXnHandover != 0 {
				xnHandoverChannel = time.After(time.Duration(simConfig.TimeBeforeXnHandover) * time.Millisecond)
			}

			var idleChannel <-chan time.Time = nil
			var reconnectChannel <-chan time.Time = nil
			if simConfig.TimeBeforeIdle != 0 {
				idleChannel = time.After(time.Duration(simConfig.TimeBeforeIdle) * time.Millisecond)
			}

			loop := true
			registered := false
			state := ueCtx.MM5G_NULL
			for loop {
				select {
				case <-deregistrationChannel:
					if ueRx != nil {
						ueRx <- procedures.UeTesterMessage{Type: procedures.Terminate}
						ueRx = nil
					}
				case <-ngapHandoverChannel:
					trigger.TriggerNgapHandover(simConfig.Gnbs[gnbIdGen(nextHandoverId)], simConfig.Gnbs[gnbIdGen(nextHandoverId+1)], int64(ueId))
					nextHandoverId++
				case <-xnHandoverChannel:
					trigger.TriggerXnHandover(simConfig.Gnbs[gnbIdGen(nextHandoverId)], simConfig.Gnbs[gnbIdGen(nextHandoverId+1)], int64(ueId))
					nextHandoverId++
				case <-idleChannel:
					if ueRx != nil {
						ueRx <- procedures.UeTesterMessage{Type: procedures.Idle}
						// Channel creation to be transformed into a task ;-)
						if simConfig.TimeBeforeReconnecting != 0 {
							reconnectChannel = time.After(time.Duration(simConfig.TimeBeforeReconnecting) * time.Millisecond)
						}
					}
				case <-reconnectChannel:
					if ueRx != nil {
						ueRx <- procedures.UeTesterMessage{Type: procedures.ServiceRequest}
					}
				case msg := <-scenarioChan:
					if ueRx != nil {
						ueRx <- msg
						if msg.Type == procedures.Terminate || msg.Type == procedures.Kill {
							ueRx = nil
						}
					}
				case msg := <-ueTx:
					log.Info("[UE] Switched from state ", state, " to state ", msg.StateChange)
					switch msg.StateChange {
					case ueCtx.MM5G_REGISTERED:
						if !registered {
							for i := 0; i < simConfig.NumPduSessions; i++ {
								ueRx <- procedures.UeTesterMessage{Type: procedures.NewPDUSession}
							}
							registered = true
						}
					case ueCtx.MM5G_NULL:
						loop = false
					}
					state = msg.StateChange
				}
			}
			if !simConfig.RegistrationLoop {
				break
			} else if simConfig.LoopCount != 0 && i == simConfig.LoopCount {
				break
			} else {
				time.Sleep(time.Duration(simConfig.TimeBeforeReregistration) * time.Millisecond)
			}
		}
	}(simConfig.ScenarioChan, simConfig.UeId)
}

func IncrementMsin(i int, msin string) string {

	msin_int, err := strconv.Atoi(msin)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given MSIN is invalid")
	}
	base := msin_int + (i - 1)

	var imsi string
	if len(msin) == 9 {
		imsi = fmt.Sprintf("%09d", base)
	} else {
		imsi = fmt.Sprintf("%010d", base)
	}
	return imsi
}
