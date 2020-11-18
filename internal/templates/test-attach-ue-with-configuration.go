package templates

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine"
	"my5G-RANTester/internal/data_test_engine"
)

func TestAttachUeWithConfiguration() error {

	cfg, err := config.GetConfig()
	if err != nil {
		return nil
	}

	fmt.Println("mytest: ", cfg.GNodeB.ControlIF.Ip, cfg.GNodeB.ControlIF.Port)

	// make N2(RAN connect to AMF)
	conn, err := control_test_engine.ConnectToAmf(cfg.AMF.Ip, cfg.GNodeB.ControlIF.Ip, cfg.AMF.Port, cfg.GNodeB.ControlIF.Port)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := data_test_engine.ConnectToUpf(cfg.GNodeB.DataIF.Ip, cfg.UPF.Ip, cfg.GNodeB.DataIF.Port, cfg.UPF.Port)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	contextGnb, err := control_test_engine.RegistrationGNB(conn, cfg.GNodeB.PlmnList.GnbId, "my5gRANTester", cfg)
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a UE.
	ue, err := control_test_engine.RegistrationUE(conn, cfg.Ue.Imsi, int64(1), cfg, contextGnb, cfg.Ue.Hplmn.Mcc, cfg.Ue.Hplmn.Mnc)
	if err != nil {
		return fmt.Errorf("The test failed when UE %s tried to attach! Error:%s", ue.Suci, err)
	}

	// data plane UE
	gtpHeader := data_test_engine.GenerateGtpHeader(int(ue.GetUeTeid()))

	// ping
	err = data_test_engine.PingUE(upfConn, gtpHeader, ue.GetIp(), cfg.Ue.Ping)
	if err != nil {
		return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	return nil
}
