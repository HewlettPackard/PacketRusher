/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package gnb

import (
	stdcontext "context"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/gtp"
	serviceNas "my5G-RANTester/internal/control_test_engine/gnb/nas/service"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
	"my5G-RANTester/internal/monitoring"

	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// InitGnb retains the original constructor signature. It consumes the caller's
// pre-added WaitGroup slot on success and failure, and logs returned errors.
// Library and CLI callers should use InitGnbContext to handle errors directly.
func InitGnb(conf config.Config, wg *sync.WaitGroup) *context.GNBContext {
	ctx, cancel := signal.NotifyContext(stdcontext.Background(), os.Interrupt)
	defer cancel()
	gnb, err := InitGnbContext(ctx, conf, wg)
	if err != nil {
		log.Error("[GNB] Startup failed: ", err)
	}
	return gnb
}

// InitGnbContext creates a gNB within ctx's total startup budget. Once returned,
// the gNB lives until Terminate or process interrupt; a startup deadline cannot
// later terminate a successfully returned node. As with InitGnb, a nonnil wg
// must contain one caller-added slot, consumed on every outcome.
func InitGnbContext(ctx stdcontext.Context, conf config.Config, wg *sync.WaitGroup) (_ *context.GNBContext, err error) {
	started := false
	defer func() {
		if !started && wg != nil {
			wg.Done()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(conf.AMFs) == 0 {
		return nil, fmt.Errorf("at least one AMF is required")
	}
	const maxRetries = 5
	const ngSetupTimeout = 2 * time.Second
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext(conf.GNodeB.PlmnList.GnbId, conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc, conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst, conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.AddrPort, conf.GNodeB.DataIF.AddrPort)
	defer func() {
		if err != nil {
			gnb.Terminate()
			gnb.WaitAssociations()
		}
	}()
	if err := gnb.ConfigureIdentity(conf.GNodeB.PlmnList.GnbIDLength, conf.GNodeB.PlmnList.CellID); err != nil {
		return nil, fmt.Errorf("invalid gNB identity: %w", err)
	}
	// Closing native associations interrupts a blocked startup write/read. Join
	// this watcher before returning, so it cannot affect a later node lifetime.
	constructionDone, watcherDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			gnb.Terminate()
		case <-constructionDone:
		}
	}()
	stopConstruction := sync.OnceFunc(func() { close(constructionDone); <-watcherDone })
	defer stopConstruction()

	// Preserve retry-consumed endpoints across every AMF and following gNB.
	currentN2IP, currentN3IP := conf.GNodeB.ControlIF, conf.GNodeB.DataIF
	for _, amfConfig := range conf.AMFs {
		if amfConfig == nil {
			return nil, fmt.Errorf("nil AMF configuration")
		}
		var lastErr error
		connected := false
		for retry := 0; retry < maxRetries; retry++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if retry > 0 {
				currentN2IP = currentN2IP.WithNextAddr()
				currentN3IP = currentN3IP.WithNextAddr()
			}
			if !currentN2IP.AddrPort.IsValid() || !currentN3IP.AddrPort.IsValid() {
				return nil, fmt.Errorf("gNB retry exhausted endpoint address range")
			}
			gnb.SetStartupEndpoints(currentN2IP.AddrPort, currentN3IP.AddrPort)
			amf := gnb.NewGnBAmf(amfConfig.AddrPort)
			if lastErr = ngap.InitConnContext(ctx, amf, gnb); lastErr == nil {
				lastErr = trigger.SendNgSetupRequest(gnb, amf)
				if lastErr == nil {
					setupCtx, cancel := stdcontext.WithTimeout(ctx, ngSetupTimeout)
					lastErr = amf.WaitActive(setupCtx, gnb.Done())
					cancel()
				}
			}
			if lastErr == nil {
				connected = true
				break
			}
			// Remove ownership before closing; a late response must not revive an
			// abandoned startup association or initiate a background redial.
			gnb.RemoveGnbAmf(amf)
			log.Warnf("[GNB] AMF %s startup attempt %d/%d failed: %v", amfConfig.AddrPort, retry+1, maxRetries, lastErr)
		}
		if !connected {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("AMF %s failed after %d startup attempts: %w", amfConfig.AddrPort, maxRetries, lastErr)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if conf.Ue.TunnelMode == config.TunnelShared && conf.Ue.TunnelBackend != config.TunnelBackendUserspace && conf.Ue.TunnelBackend != config.TunnelBackendEBPF {
		dev, err := gtp.NewDevice(gnb.GetN3GnbIp(), conf.Ue.TunnelMTU)
		if err != nil {
			return nil, fmt.Errorf("create shared GTP-U device: %w", err)
		}
		if !gnb.PublishGtpDevice(dev) {
			dev.Close()
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return nil, stdcontext.Canceled
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stopConstruction()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	serviceNas.InitServer(gnb)
	sigGnb := make(chan os.Signal, 1)
	signal.Notify(sigGnb, os.Interrupt)
	started = true
	go func() {
		defer signal.Stop(sigGnb)
		if wg != nil {
			defer wg.Done()
		}
		select {
		case <-sigGnb:
			gnb.Terminate()
		case <-gnb.Done():
		}
		gnb.WaitAssociations()
	}()
	return gnb, nil
}

// Monitoring probes keep their original one-second NG Setup observation
// window, but share the owned cancellable transport and error handling.
func InitGnbForLoadSeconds(conf config.Config, wg *sync.WaitGroup, monitor *monitoring.Monitor) {
	defer wg.Done()
	ctx, cancel := signal.NotifyContext(stdcontext.Background(), os.Interrupt)
	defer cancel()
	if err := probeAMFs(ctx, conf, monitor.IncRqs); err != nil {
		log.Error("[GNB] Load probe failed: ", err)
	}
}

func InitGnbForAvaibility(conf config.Config, monitor *monitoring.Monitor) {
	ctx, cancel := signal.NotifyContext(stdcontext.Background(), os.Interrupt)
	defer cancel()
	if err := probeAMFs(ctx, conf, monitor.IncAvaibility); err != nil {
		log.Error("[GNB] Availability probe failed: ", err)
	}
}

func probeAMFs(ctx stdcontext.Context, conf config.Config, observed func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext(conf.GNodeB.PlmnList.GnbId, conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc, conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst, conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.AddrPort, conf.GNodeB.DataIF.AddrPort)
	defer func() { gnb.Terminate(); gnb.WaitAssociations() }()
	if err := gnb.ConfigureIdentity(conf.GNodeB.PlmnList.GnbIDLength, conf.GNodeB.PlmnList.CellID); err != nil {
		return fmt.Errorf("invalid gNB identity: %w", err)
	}
	watcherDone, finished := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			gnb.Terminate()
		case <-finished:
		}
	}()
	defer func() { close(finished); <-watcherDone }()
	for _, config := range conf.AMFs {
		if config == nil {
			return fmt.Errorf("nil AMF configuration")
		}
		amf := gnb.NewGnBAmf(config.AddrPort)
		if err := ngap.InitConnContext(ctx, amf, gnb); err != nil {
			return err
		}
		if err := trigger.SendNgSetupRequest(gnb, amf); err != nil {
			return err
		}
		readyCtx, cancel := stdcontext.WithTimeout(ctx, time.Second)
		err := amf.WaitActive(readyCtx, gnb.Done())
		cancel()
		if err != nil {
			return err
		}
		observed()
	}
	return nil
}
