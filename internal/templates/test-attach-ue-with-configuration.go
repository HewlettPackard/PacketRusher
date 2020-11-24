package templates

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine"
	"my5G-RANTester/internal/data_test_engine"
)

func TestAttachUeWithConfiguration() {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Error in get configuration")
	}

	//fmt.Println("mytest: ", cfg.GNodeB.ControlIF.Ip, cfg.GNodeB.ControlIF.Port)
	log.Info(fmt.Sprintf("[CORE]%s Core in Testing\n", cfg.AMF.Name))

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(cfg.AMF.Ip, cfg.GNodeB.ControlIF.Ip, cfg.AMF.Port, cfg.GNodeB.ControlIF.Port)
	if err != nil {
		log.Fatal("The test failed when sctp socket tried to connect to AMF! Error:", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := data_test_engine.ConnectToUpf(cfg.GNodeB.DataIF.Ip, cfg.UPF.Ip, cfg.GNodeB.DataIF.Port, cfg.UPF.Port)
	if err != nil {
		log.Fatal("The test failed when udp socket tried to connect to UPF! Error:", err)
	}

	// authentication to a GNB.
	contextGnb, err := control_test_engine.RegistrationGNB(conn, cfg.GNodeB.PlmnList.GnbId, "my5gRANTester", cfg)
	if err != nil {
		log.Fatal("The test failed when GNB tried to attach! Error:", err)
	}

	// authentication to a UE.
	ue, err := control_test_engine.RegistrationUE(conn, cfg.Ue.Imsi, int64(1), cfg, contextGnb, cfg.Ue.Hplmn.Mcc, cfg.Ue.Hplmn.Mnc)
	if err != nil {
		log.Error("The test failed when UE", ue.Suci, "tried to attach! Error:", err)
	}

	// data plane UE
	gtpHeader := data_test_engine.GenerateGtpHeader(int(ue.GetUeTeid()))

	// ping
	err = data_test_engine.PingUE(upfConn, gtpHeader, ue.GetIp(), cfg.Ue.Ping)
	if err != nil {
		log.Error("The test failed when UE tried to use ping! Error:", err)
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	// return nil
}
