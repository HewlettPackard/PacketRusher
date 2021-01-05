package gnb

import (
	"fmt"
	"log"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	serviceNas "my5G-RANTester/internal/control_test_engine/gnb/nas/service"
	serviceNgap "my5G-RANTester/internal/control_test_engine/gnb/ngap/service"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
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
		conf.GNodeB.SliceSupportList.Sst,
		conf.GNodeB.ControlIF.Ip,
		conf.GNodeB.ControlIF.Port)

	// start communication with AMF(server SCTP).

	// new AMF context.
	amf := gnb.NewGnBAmf(conf.AMF.Ip, conf.AMF.Port)

	if err := serviceNgap.InitConn(amf, gnb); err != nil {
		log.Fatal("Error in", err)
	} else {
		fmt.Println("[GNB] SCTP/NGAP service is running")
		wg.Add(1)
	}

	// start communication with UE(server UNIX sockets).
	if err := serviceNas.InitServer(gnb); err != nil {
		log.Fatal("Error in", err)
	} else {
		fmt.Println("[GNB] UNIX/NAS service is running")
		wg.Add(1)
	}

	trigger.SendNgSetupRequest(gnb, amf)

	wg.Wait()
}
