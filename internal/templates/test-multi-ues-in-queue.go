/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package templates

import (
	"context"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, tunnelBackend config.TunnelBackend, dedicatedGnb bool, loop bool, loopCount int, timeBeforeReregistration int, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeReconnecting int, numPduSessions int, numGnbs int, controlSocket string) error {
	if tunnelMode != config.TunnelDisabled {
		if !dedicatedGnb && tunnelMode != config.TunnelShared {
			log.Fatal("You cannot use the --tunnel option, without using the --dedicatedGnb option")
		}
		if timeBetweenRegistration < 500 && tunnelMode != config.TunnelShared && tunnelBackend == config.TunnelBackendGtp5g {
			log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
		}
	}

	if numPduSessions < 0 || numPduSessions > 15 {
		log.Fatal("Each UE can have 0 to 15 PDU Sessions (NAS PDU session identities 1 to 15).")
	}

	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	// Set before the gNBs are created: in shared mode each gNB creates the GTP-U
	// device its UEs will use.
	cfg.Ue.TunnelMode = tunnelMode
	cfg.Ue.TunnelBackend = tunnelBackend
	if tunnelBackend != config.TunnelBackendGtp5g && tunnelMode == config.TunnelShared {
		// Only gtp5g shares a device: with the others each UE of the gNB gets its own, as with -d.
		cfg.Ue.TunnelMode = config.TunnelTun
	}

	var numGnb int
	if dedicatedGnb {
		numGnb = numUes
	} else {
		numGnb = 1
	}
	if numGnbs > numGnb {
		numGnb = numGnbs
	}
	if numGnb <= 1 && (timeBeforeXnHandover != 0 || timeBeforeNgapHandover != 0) {
		log.Warn("[TESTER] We are increasing the number of gNodeB to two for handover test cases. Make you sure you fill the requirements for having two gNodeBs.")
		numGnb++
	}

	// Ctrl-C interrupts the gNBs while they start, and stops the UEs once they run.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	gnbs, err := tools.CreateGnbs(ctx, numGnb, cfg, &wg)
	if err != nil {
		return err
	}

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(1 * time.Second)

	var controlServer *control.Server
	if controlSocket != "" {
		var err error
		controlServer, err = control.Listen(controlSocket, gnbs, numUes)
		if err != nil {
			return fmt.Errorf("unable to create the control socket: %w", err)
		}
		defer controlServer.Close()
	}

	simulations := make([]*tools.UESimulation, 0, numUes)

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
		simulation := tools.SimulateSingleUE(ueSimCfg, &wg)
		simulations = append(simulations, simulation)
		if controlServer != nil {
			controlServer.AddUe(ueSimCfg.UeId, simulation)
		}

		// Before creating a new UE, we wait for timeBetweenRegistration ms
		registrationDelay := time.NewTimer(time.Duration(timeBetweenRegistration) * time.Millisecond)
		select {
		case <-ctx.Done():
			registrationDelay.Stop()
			stopSignal = false
		case <-registrationDelay.C:
		}
	}

	if stopSignal {
		<-ctx.Done()
	}
	stopUESimulations(simulations)

	// Each scenario has finished local UE cleanup before its shared device closes.
	// Keep the existing bounds for any residual holds from interrupted procedures.
	for _, gnb := range gnbs {
		gnb.CloseGtpDevice(5*time.Second, 60*time.Second)
	}
	return nil
}

// Completed scenarios no longer have a command receiver. Send checks their
// completion, and remains blocking for live scenarios so none misses shutdown.
func stopUESimulations(simulations []*tools.UESimulation) {
	for _, simulation := range simulations {
		simulation.Send(procedures.UeTesterMessage{Type: procedures.Terminate})
	}
	for _, simulation := range simulations {
		<-simulation.Done()
	}
}
