package templates

import (
	"encoding/hex"
	"fmt"
	"my5G-RANTester/internal/control_test_engine"
	"sync"
)

func TestMultiAttachGnbInConcurrency(numberGnbs int) error {

	ranPort := 9487
	var wg sync.WaitGroup

	// multiple concurrent GNBs authentication using goroutines.
	for i := 1; i <= numberGnbs; i++ {

		wg.Add(1)
		go func(wg *sync.WaitGroup, ranPort int, i int) {

			defer wg.Done()

			// make N2(RAN connect to AMF)
			conn, err := control_test_engine.ConnectToAmf("127.0.0.1", "127.0.0.1", 38412, ranPort)
			if err != nil {
				fmt.Printf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
			}

			// multiple names for GNBs.
			nameGNB := "my5gRanTester" + string(i)

			// generate GNB id.
			var aux string
			if i < 16 {
				aux = "00000" + fmt.Sprintf("%x", i)
			} else if i < 256 {
				aux = "0000" + fmt.Sprintf("%x", i)
			} else {
				aux = "000" + fmt.Sprintf("%x", i)
			}

			resu, err := hex.DecodeString(aux)
			if err != nil {
				fmt.Printf("error in GNB id for testing multiple GNBs")
			}

			// authentication to a GNB.
			err = control_test_engine.RegistrationGNB(conn, resu, nameGNB)
			if err != nil {
				fmt.Printf("The test failed when GNB tried to attach! Error:%s", err)
			}
		}(&wg, ranPort, i)
		ranPort++
	}

	// wait threads.
	wg.Wait()

	return nil
}
