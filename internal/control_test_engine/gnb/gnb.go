/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package gnb

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
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

func InitGnb(conf config.Config, wg *sync.WaitGroup) *context.GNBContext {

	// instance new gnb.
	gnb := &context.GNBContext{}

	const maxRetries = 5
	const ngSetupTimeout = 2 * time.Second

	log.Info("[GNB] InitGnb called with N2 IP: ", conf.GNodeB.ControlIF.String())

	// start communication with AMF (server SCTP).
	for _, amfConfig := range conf.AMFs {
		connected := false
		currentN2IP := conf.GNodeB.ControlIF

		for retry := 0; retry < maxRetries && !connected; retry++ {
			if retry > 0 {
				// Increment N2 IP address for retry
				currentN2IP = currentN2IP.WithNextAddr()
				log.Info("[GNB] Retrying with N2 IP: ", currentN2IP.String())
			} else {
				log.Info("[GNB] Attempting connection with N2 IP: ", currentN2IP.String())
			}

			// Initialize/re-initialize gnb context with current IP
			gnb.NewRanGnbContext(
				conf.GNodeB.PlmnList.GnbId,
				conf.GNodeB.PlmnList.Mcc,
				conf.GNodeB.PlmnList.Mnc,
				conf.GNodeB.PlmnList.Tac,
				conf.GNodeB.SliceSupportList.Sst,
				conf.GNodeB.SliceSupportList.Sd,
				currentN2IP.AddrPort,
				conf.GNodeB.DataIF.AddrPort,
			)

			// new AMF context.
			amf := gnb.NewGnBAmf(amfConfig.AddrPort)

			// start communication with AMF(SCTP).
			if err := ngap.InitConn(amf, gnb); err != nil {
				log.Warn("[GNB] Failed to connect to AMF (attempt ", retry+1, "/", maxRetries, "): ", err)
				if retry == maxRetries-1 {
					log.Fatal("[GNB] Max retries reached, failed to connect to AMF")
				}
				continue
			}

			log.Info("[GNB] SCTP/NGAP service is running")

			// Send NG Setup Request
			log.Info("[GNB] Sending NG Setup Request...")
			trigger.SendNgSetupRequest(gnb, amf)
			log.Info("[GNB] NG Setup Request sent, waiting for response...")

			// Wait for NG Setup Response with timeout
			setupSuccess := false
			timeoutChan := time.After(ngSetupTimeout)
			checkInterval := time.NewTicker(100 * time.Millisecond)

		waitLoop:
			for !setupSuccess {
				select {
				case <-timeoutChan:
					log.Warn("[GNB] NG Setup timeout after ", ngSetupTimeout, " (attempt ", retry+1, "/", maxRetries, "), AMF state: ", amf.GetState())
					// Close SCTP connection
					if conn := amf.GetSCTPConn(); conn != nil {
						log.Info("[GNB] Closing SCTP connection...")
						conn.Close()
					}
					checkInterval.Stop()
					break waitLoop
				case <-checkInterval.C:
					// Check AMF state
					currentState := amf.GetState()
					if currentState == 0x01 { // Active state
						log.Info("[GNB] NG Setup successful with N2 IP: ", currentN2IP.String())
						connected = true
						setupSuccess = true
						checkInterval.Stop()
						break waitLoop
					}
				}
			}

			if connected {
				break
			}
		}

		if !connected {
			log.Fatal("[GNB] Failed to establish connection after ", maxRetries, " retries")
		}
	}

	// start communication with UE (server UNIX sockets).
	serviceNas.InitServer(gnb)

	go func() {
		// control the signals
		sigGnb := make(chan os.Signal, 1)
		signal.Notify(sigGnb, os.Interrupt)

		// Block until a signal is received.
		<-sigGnb
		//gnb.Terminate()
		wg.Done()
	}()

	return gnb
}

func InitGnbForLoadSeconds(conf config.Config, wg *sync.WaitGroup,
	monitor *monitoring.Monitor) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.AddrPort,
		conf.GNodeB.DataIF.AddrPort)

	// start communication with AMF (server SCTP).
	for _, amf := range conf.AMFs {
		// new AMF context.
		amf := gnb.NewGnBAmf(amf.AddrPort)

		// start communication with AMF(SCTP).
		ngap.InitConn(amf, gnb)

		trigger.SendNgSetupRequest(gnb, amf)

		// timeout is 1 second for receive NG Setup Response
		time.Sleep(1000 * time.Millisecond)

		// AMF responds message sends by Tester
		// means AMF is available
		if amf.GetState() == 0x01 {
			monitor.IncRqs()
		}
	}

	gnb.Terminate()
	wg.Done()
	// os.Exit(0)
}

func InitGnbForAvaibility(conf config.Config,
	monitor *monitoring.Monitor) {

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.SliceSupportList.Sd,
		conf.GNodeB.ControlIF.AddrPort,
		conf.GNodeB.DataIF.AddrPort)

	// start communication with AMF (server SCTP).
	for _, amf := range conf.AMFs {
		// new AMF context.
		amf := gnb.NewGnBAmf(amf.AddrPort)

		// start communication with AMF(SCTP).
		if err := ngap.InitConn(amf, gnb); err != nil {
			log.Info("Error in ", err)

			return

		} else {
			log.Info("[GNB] SCTP/NGAP service is running")

		}

		trigger.SendNgSetupRequest(gnb, amf)

		// timeout is 1 second for receive NG Setup Response
		time.Sleep(1000 * time.Millisecond)

		// AMF responds message sends by Tester
		// means AMF is available
		if amf.GetState() == 0x01 {
			monitor.IncAvaibility()

		}
	}

	gnb.Terminate()
	// os.Exit(0)
}
