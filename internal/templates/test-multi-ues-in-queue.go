package templates

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

func TestMultiUesInQueue(numUes int, tunnelEnabled bool, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int) {
	if tunnelEnabled && !dedicatedGnb {
		log.Fatal("You cannot use the --tunnel option, without using the --dedicatedGnb option")
	}

	if tunnelEnabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("[TESTER][CONFIG] Unable to read configuration")
	}

	var numGnb int
	if dedicatedGnb {
		numGnb = numUes
	} else {
		numGnb = 1
	}

	// Each gNB have their own IP address on both N2 and N3
	// TODO: Limitation for now, these IPs must be sequential, eg:
	// gnb[0].n2_ip = 192.168.2.10, gnb[0].n3_ip = 192.168.3.10
	// gnb[1].n2_ip = 192.168.2.11, gnb[1].n3_ip = 192.168.3.11
	// ...
	n2Ip := cfg.GNodeB.ControlIF.Ip
	n3Ip := cfg.GNodeB.DataIF.Ip
	for i := 1; i <= numGnb; i++ {
		cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)
		cfg.GNodeB.ControlIF.Ip = n2Ip
		cfg.GNodeB.DataIF.Ip = n3Ip

		go gnb.InitGnb(cfg, &wg)
		wg.Add(1)

		// TODO: We could find the interfaces where N2/N3 are
		// and check that the generated IPs, still belong to the interfaces' subnet
		n2Ip, _ = incrementIP(n2Ip, "0.0.0.0/0")
		n3Ip, _ = incrementIP(n3Ip, "0.0.0.0/0")
	}

	// Wait for gNB to be connected before registering UEs
	// TODO: We should wait for NGSetupResponse instead
	time.Sleep(1 * time.Second)

	msin := cfg.Ue.Msin
	cfg.Ue.TunnelEnabled = tunnelEnabled

	ues := make([]chan procedures.UeTesterMessage, numUes+1)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	stopSignal := true
	for stopSignal {
		// If CTRL-C signal has been received,
		// stop creating new UEs, else we create numUes UEs
		for i := 1; stopSignal && i <= numUes; i++ {
			imsi := imsiGenerator(i, msin)
			log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", imsi, " UE")
			cfg.Ue.Msin = imsi

			// Associate UE[i] with gnb[i] when dedicatedGnb = true
			// else all UE[i] are associated with gnb[0]
			if dedicatedGnb {
				cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)
			}

			// If there is currently a coroutine handling current UE
			// kill it, before creating a new coroutine with same UE
			// Use case: Registration of N UEs in loop, when loop = true
			if ues[i] != nil {
				ues[i] <- procedures.Kill
			}
			ues[i] = make(chan procedures.UeTesterMessage)

			// Launch a coroutine to handle UE
			wg.Add(1)
			go func() {
				ueChan := ues[i]

				// Create a new UE coroutine
				// ue.NewUE returns context of the new UE
				ue := ue.NewUE(cfg, uint8(i), ueChan, &wg)

				// We tell the UE to perform a registration
				ueChan <- procedures.Registration
				for {
					// TODO: Add timeout + check for unexpected state
					// When the UE is registered, tell the UE to trigger a PDU Session
					if ue.WaitOnStateMM() == context.MM5G_REGISTERED {
						ueChan <- procedures.NewPDUSession
						break
					}
				}

				if timeBeforeDeregistration > 0 {
					time.Sleep(time.Duration(timeBeforeDeregistration) * time.Millisecond)
					ueChan <- procedures.Deregistration
				}
			}()

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
	wg.Wait()
}

func imsiGenerator(i int, msin string) string {

	msin_int, err := strconv.Atoi(msin)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given MSIN is invalid")
	}
	base := msin_int + (i - 1)

	imsi := fmt.Sprintf("%010d", base)
	return imsi
}

func incrementIP(origIP, cidr string) (string, error) {
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
		log.Fatal("[GNB][CONFIG] gNB IP Address is not in N2/N3 subnet")
	}
	return ip.String(), nil
}
