package templates

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine"
	"sync"
)

func TestMultiAttachGnbInConcurrency(numberGnbs int) error {

	var wg sync.WaitGroup

	cfg, err := config.GetConfig()
	if err != nil {
		return nil
	}

	ranPort := cfg.GNodeB.ControlIF.Port

	// multiple concurrent GNBs authentication using goroutines.
	for i := 1; i <= numberGnbs; i++ {

		wg.Add(1)
		go func(wg *sync.WaitGroup, ranPort int, i int) {

			defer wg.Done()

			// make N2(RAN connect to AMF)
			conn, err := control_test_engine.ConnectToAmf(cfg.AMF.Ip, cfg.GNodeB.ControlIF.Ip, cfg.AMF.Port, ranPort)
			if err != nil {
				fmt.Printf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
			}

			// multiple names for GNBs.
			nameGNB := fmt.Sprint("my5gRanTester", i)
			// fmt.Println(nameGNB)

			// generate GNB id.
			var aux string
			if i < 16 {
				aux = "00000" + fmt.Sprintf("%x", i)
			} else if i < 256 {
				aux = "0000" + fmt.Sprintf("%x", i)
			} else {
				aux = "000" + fmt.Sprintf("%x", i)
			}

			/*
				resu, err := hex.DecodeString(aux)
				if err != nil {
					fmt.Printf("error in GNB id for testing multiple GNBs")
				}
			*/

			// authentication to a GNB.
			err = control_test_engine.RegistrationGNB(conn, aux, nameGNB, cfg)
			if err != nil {
				fmt.Printf("The test failed when GNB tried to attach! Error:%s", err)
			}

			// close sctp socket.
			conn.Close()
		}(&wg, ranPort, i)
		ranPort++
	}

	// wait threads.
	wg.Wait()

	return nil
}
