package templates

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
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

	n2Ip := cfg.GNodeB.ControlIF.Ip
	n3Ip := cfg.GNodeB.DataIF.Ip
	for i := 1; i <= numGnb; i++ {
		cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)
		cfg.GNodeB.ControlIF.Ip = n2Ip
		cfg.GNodeB.DataIF.Ip = n3Ip

		go gnb.InitGnb(cfg, &wg)
		wg.Add(1)

		n2Ip, _ = incrementIP(n2Ip, "0.0.0.0/0")
		n3Ip, _ = incrementIP(n3Ip, "0.0.0.0/0")
	}

	time.Sleep(1 * time.Second)

	msin := cfg.Ue.Msin
	cfg.Ue.TunnelEnabled = tunnelEnabled

	ues := make([]chan string, numUes+1)

	sigStop := make(chan os.Signal, 1)
	signal.Notify(sigStop, os.Interrupt)

	stopSignal := true
	for stopSignal {
		for i := 1; stopSignal && i <= numUes; i++ {
			imsi := imsiGenerator(i, msin)
			log.Info("[TESTER] TESTING REGISTRATION USING IMSI ", imsi, " UE")
			cfg.Ue.Msin = imsi
			if dedicatedGnb {
				cfg.GNodeB.PlmnList.GnbId = gnbIdGenerator(i)
			}

			if ues[i] != nil {
				ues[i] <- "kill"
			}
			ues[i] = make(chan string)

			go ue.RegistrationUe(cfg, uint8(i), ues[i], timeBeforeDeregistration, &wg)

			wg.Add(1)

			time.Sleep(time.Duration(timeBetweenRegistration) * time.Millisecond)
			select {
			case <-sigStop:
				stopSignal = false
			default:
			}
		}
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
