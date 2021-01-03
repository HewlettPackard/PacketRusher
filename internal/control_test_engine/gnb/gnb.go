package gnb

import (
	"fmt"
	"log"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/service"
	"sync"
)

func InitGnb(conf config.Config) {

	wg := sync.WaitGroup{}

	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		conf.GNodeB.PlmnList.GnbId,
		conf.GNodeB.PlmnList.Mcc,
		conf.GNodeB.PlmnList.Mnc,
		conf.GNodeB.PlmnList.Tac,
		conf.GNodeB.SliceSupportList.St,
		conf.GNodeB.SliceSupportList.Sst)

	// start communication with AMF(server SCTP).

	// start communication with UE(server UNIX sockets).
	if err := service.InitServer(gnb); err != nil {
		log.Fatal("Error in", err)
	} else {
		fmt.Println("[GNB] Unix sockets service is running")
		wg.Add(1)
	}

	wg.Wait()
}
