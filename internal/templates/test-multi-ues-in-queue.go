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
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, loopCount int, timeBeforeReregistration int, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeReconnecting int, numPduSessions int, churnFirstN int) {
	if tunnelMode != config.TunnelDisabled {
		if !dedicatedGnb && tunnelMode != config.TunnelShared {
			log.Fatal("You cannot use the --tunnel option, without using the --dedicatedGnb option")
		}
		if timeBetweenRegistration < 500 && tunnelMode != config.TunnelShared {
			log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
		}
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cohort := newChurnCohort(churnFirstN, numUes, time.Duration(timeBetweenRegistration)*time.Millisecond)

	// Subscribed before anything starts. Without a subscription Go ignores SIGUSR1
	// and SIGUSR2, so a wave signalled during start-up or the fill would be lost; with
	// it, the signal is queued and its wave runs once the fill is done. The queue holds
	// eight signals; signal.Notify drops any beyond that.
	sigChurn := make(chan os.Signal, 8)
	if cohort != nil {
		signal.Notify(sigChurn, syscall.SIGUSR1, syscall.SIGUSR2)
	}

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// Set before the gNBs are created: in shared mode each gNB creates the GTP-U
	// device its UEs will use.
	cfg.Ue.TunnelMode = tunnelMode

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
		RegistrationLoop:         loop,
		LoopCount:                loopCount,
		TimeBeforeReregistration: timeBeforeReregistration,
	}

	stopSignal := true
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

		// ueSimCfg is reused for every UE, so each field is set either way: left over,
		// it would put every UE after the cohort in it too. A cohort member leaves only
		// on a wave: its own deregistration timer would park it behind the driver's
		// back, and the next wave's Terminate would then end it for good.
		ueSimCfg.Churn = cohort.memberFor(ueSimCfg.UeId)
		if ueSimCfg.Churn != nil {
			ueSimCfg.RegistrationLoop = true
			ueSimCfg.LoopCount = 0
			ueSimCfg.TimeBeforeDeregistration = 0
		} else {
			ueSimCfg.RegistrationLoop = loop
			ueSimCfg.LoopCount = loopCount
			ueSimCfg.TimeBeforeDeregistration = timeBeforeDeregistration
		}

		tools.SimulateSingleUE(ueSimCfg, &wg)

		// Before creating a new UE, we wait for timeBetweenRegistration ms
		time.Sleep(time.Duration(timeBetweenRegistration) * time.Millisecond)

		select {
		case <-sigStop:
			stopSignal = false
		default:
		}
	}

	if stopSignal && cohort != nil {
		log.Info("[TESTER][CHURN] fill complete; the cohort is UE 1..", cohort.size, " of ", numUes,
			". SIGUSR1 deregisters it, SIGUSR2 registers it again, ", timeBetweenRegistration,
			" ms apart. SIGINT stops.")
		for waves := true; waves; {
			select {
			case <-sigStop:
				waves = false
			case sig := <-sigChurn:
				if sig == syscall.SIGUSR1 {
					waves = !cohort.waveOut(func(id int) {
						sendUnlessDone(scenarioChans[id], procedures.UeTesterMessage{Type: procedures.Terminate}, cohort.memberFor(id))
					}, sigStop)
				} else {
					waves = !cohort.waveIn(sigStop)
				}
			}
		}
	} else if stopSignal {
		<-sigStop
	}
	for id, scenarioChan := range scenarioChans {
		if scenarioChan != nil {
			sendUnlessDone(scenarioChan, procedures.UeTesterMessage{Type: procedures.Terminate}, cohort.memberFor(id))
		}
	}

	time.Sleep(time.Second * 1)

	// Each gNB removes the GTP-U device its UEs shared, once they have released their
	// tunnels on it. It stops waiting when releases stop coming for 5 s, or after 60 s:
	// UEs that cannot reach the AMF never release, and in --loop mode UEs register
	// again after they terminate.
	for _, gnb := range gnbs {
		gnb.CloseGtpDevice(5*time.Second, 60*time.Second)
	}
}

// sendUnlessDone sends msg to a UE's scenario channel. For a churn cohort member it
// gives up once the member has returned, since nothing reads the channel then; other
// UEs are sent to as before.
func sendUnlessDone(ch chan procedures.UeTesterMessage, msg procedures.UeTesterMessage, member *tools.ChurnMember) {
	if member == nil {
		ch <- msg
		return
	}
	select {
	case ch <- msg:
	case <-member.Done():
	}
}
