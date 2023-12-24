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
	var err error
	// Each gNB have their own IP address on both N2 and N3
	// TODO: Limitation for now, these IPs must be sequential, eg:
	// gnb[0].n2_ip = 192.168.2.10, gnb[0].n3_ip = 192.168.3.10
	// gnb[1].n2_ip = 192.168.2.11, gnb[1].n3_ip = 192.168.3.11
	// ...
	gnbId := cfg.GNodeB.PlmnList.GnbId
	n2Ip := cfg.GNodeB.ControlIF.Ip
	n3Ip := cfg.GNodeB.DataIF.Ip
	for i := 1; i <= count; i++ {
		cfg.GNodeB.PlmnList.GnbId = gnbId
		cfg.GNodeB.ControlIF.Ip = n2Ip
		cfg.GNodeB.DataIF.Ip = n3Ip

		gnbs[cfg.GNodeB.PlmnList.GnbId] = gnb.InitGnb(cfg, wg)
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the generated IPs, still belong to the interfaces' subnet
		gnbId = gnbIdGenerator(i, gnbId)
		n2Ip, err = IncrementIP(n2Ip, "0.0.0.0/0")
		if err != nil {
			log.Fatal("[GNB][CONFIG] Error while allocating ip for N2: " + err.Error())
		}
		n3Ip, err = IncrementIP(n3Ip, "0.0.0.0/0")
		if err != nil {
			log.Fatal("[GNB][CONFIG] Error while allocating ip for N3: " + err.Error())
		}
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

	gnbId = fmt.Sprintf("%06x", base)
	return gnbId
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

type UESimulationConfig struct {
	UeId                     int
	Gnbs                     map[string]*gnbCxt.GNBContext
	Cfg                      config.Config
	ScenarioChan             chan procedures.UeTesterMessage
	TimeBeforeDeregistration int
	TimeBeforeHandover       int
	NumPduSessions           int
}

func SimulateSingleUE(simConfig UESimulationConfig, wg *sync.WaitGroup) {
	numGnb := len(simConfig.Gnbs)
	ueCfg := simConfig.Cfg
	ueCfg.Ue.Msin = IncrementMsin(simConfig.UeId, simConfig.Cfg.Ue.Msin)
	log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", ueCfg.Ue.Msin, " UE")

	gnbId := gnbIdGenerator(simConfig.UeId%numGnb, ueCfg.GNodeB.PlmnList.GnbId)

	// Launch a coroutine to handle UE's individual scenario
	go func(scenarioChan chan procedures.UeTesterMessage, ueId int) {
		wg.Add(1)

		ueRx := make(chan procedures.UeTesterMessage)

		// Create a new UE coroutine
		// ue.NewUE returns context of the new UE
		ueTx := ue.NewUE(ueCfg, ueId, ueRx, simConfig.Gnbs[gnbId], wg)

		// We tell the UE to perform a registration
		ueRx <- procedures.UeTesterMessage{Type: procedures.Registration}

		var deregistrationChannel <-chan time.Time = nil
		if simConfig.TimeBeforeDeregistration != 0 {
			deregistrationChannel = time.After(time.Duration(simConfig.TimeBeforeDeregistration) * time.Millisecond)
		}
		var handoverChannel <-chan time.Time = nil
		if simConfig.TimeBeforeHandover != 0 {
			handoverChannel = time.After(time.Duration(simConfig.TimeBeforeHandover) * time.Millisecond)
		}

		loop := true
		state := ueCtx.MM5G_NULL
		for loop {
			select {
			case <-deregistrationChannel:
				if ueRx != nil {
					ueRx <- procedures.UeTesterMessage{Type: procedures.Terminate}
					ueRx = nil
				}
			case <-handoverChannel:
				trigger.TriggerHandover(simConfig.Gnbs[gnbId], simConfig.Gnbs[gnbIdGenerator((ueId+1)%numGnb, ueCfg.GNodeB.PlmnList.GnbId)], int64(ueId))
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
					if state != msg.StateChange {
						for i := 0; i < simConfig.NumPduSessions; i++ {
							ueRx <- procedures.UeTesterMessage{Type: procedures.NewPDUSession}
						}
					}
				case ueCtx.MM5G_NULL:
					loop = false
				}
				state = msg.StateChange
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

	imsi := fmt.Sprintf("%010d", base)
	return imsi
}
