package templates

import (
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"sync"
	"time"
)

func TestAttachUeWithConfiguration(tunnelEnabled bool) {

	wg := sync.WaitGroup{}

	cfg, err := config.GetConfig()
	if err != nil {
		//return nil
		log.Fatal("Error in get configuration")
	}

	cfg.Ue.TunnelEnabled = tunnelEnabled

	wg.Add(1)

	go gnb.InitGnb(cfg, &wg)

	time.Sleep(1 * time.Second)

	ueChan := make(chan procedures.UeTesterMessage)

	wg.Add(1)

	// Launch a coroutine for procedures handling
	go func() {
		// Create a new UE coroutine
		// ue.NewUE returns context of the new UE
		ue := ue.NewUE(cfg, 1, ueChan, &wg)
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
	}()

	wg.Wait()
}
