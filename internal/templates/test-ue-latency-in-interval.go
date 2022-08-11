package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/ue"
	"my5G-RANTester/internal/monitoring"
	"sync"
	"time"
)

// gera uma UE registration e mede a latência
func TestUesLatencyInInterval(interval int) int64 {

	wg := sync.WaitGroup{}

	monitor := monitoring.Monitor{
		LtRegisterGlobal: 0,
	}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	// creates an sctp socket (GNB) for each UE
	for i := 1; i <= interval; i++ {

		// sinal da gnb para manter a execução
		sigGnb := make(chan bool, 1)
		synch := make(chan bool, 1)

		log.Warn("[TESTER][UE] Test UE REGISTRATION:")

		// start the time
		start := time.Now()

		// usado para sincronizar se a thread da gnb gerou um erro
		go gnb.InitGnbForUeLatency(cfg, sigGnb, synch)

		// não houve erro na gnb
		if <-synch {

			time.Sleep(400 * time.Millisecond)

			go ue.RegistrationUeMonitor(cfg, uint8(i), &monitor, &wg, start)

			wg.Add(1)

			wg.Wait()

			// increment the latency global for the mean
			monitor.SetLtGlobal(monitor.LtRegisterLocal)

		} else {

			log.Warn("[TESTER][UE] UE LATENCY IN REGISTRATION: WITHOUT CONNECTION")
		}

		time.Sleep(600 * time.Millisecond)

		// seta o sinal e termina a gnb
		sigGnb <- true

		time.Sleep(40 * time.Millisecond)
	}

	return monitor.LtRegisterGlobal
}
