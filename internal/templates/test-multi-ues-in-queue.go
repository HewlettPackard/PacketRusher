/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/scenario"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ControlOptions struct {
	Socket       string
	NumberOfGnbs int
}

func TestMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, loopCount int, timeBeforeReregistration int, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeReconnecting int, numPduSessions int, controlOptions ...ControlOptions) error {
	var options ControlOptions
	if len(controlOptions) != 0 {
		options = controlOptions[0]
	}
	if options.NumberOfGnbs < 0 || options.NumberOfGnbs > 0 && dedicatedGnb {
		return fmt.Errorf("--number-of-gnbs requires a non-dedicated gNB configuration")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if tunnelMode != config.TunnelDisabled {
		if !dedicatedGnb && tunnelMode != config.TunnelShared {
			log.Fatal("You cannot use the --tunnel option, without using the --dedicatedGnb option")
		}
		if timeBetweenRegistration < 500 && tunnelMode != config.TunnelShared && config.GetConfig().Ue.TunnelBackend != config.TunnelBackendUserspace {
			log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
		}
	}

	if numPduSessions < 0 || numPduSessions > 15 {
		return fmt.Errorf("each UE requires 0 to 15 PDU sessions (0 for registration-only)")
	}
	if numPduSessions == 0 && tunnelMode != config.TunnelDisabled {
		return fmt.Errorf("a tunnel requires at least one PDU session")
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
	if options.NumberOfGnbs > 0 {
		numGnb = options.NumberOfGnbs
	}
	if numGnb <= 1 && (timeBeforeXnHandover != 0 || timeBeforeNgapHandover != 0) {
		log.Warn("[TESTER] We are increasing the number of gNodeB to two for handover test cases. Make you sure you fill the requirements for having two gNodeBs.")
		numGnb++
	}
	gnbs := tools.CreateGnbs(numGnb, cfg, &wg)
	defer func() {
		for _, node := range gnbs {
			node.Terminate()
		}
	}()
	registry := scenario.NewRegistry(gnbs)
	var controlServer *scenario.Server
	if options.Socket != "" {
		server, err := scenario.Listen(options.Socket, registry)
		if err != nil {
			return err
		}
		controlServer = server
		defer server.Close()
	}

	// A connected SCTP socket does not mean the AMF accepted NG Setup.
	readyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err := tools.WaitGnbs(readyCtx, gnbs)
	cancel()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
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
		registry.Add(ueSimCfg.UeId, simulation)

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
	if controlServer != nil {
		_ = controlServer.Close()
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
