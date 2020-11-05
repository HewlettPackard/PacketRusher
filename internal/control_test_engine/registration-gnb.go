package control_test_engine

import (
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/context"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
	"my5G-RANTester/internal/logging"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId string, nameGNB string, conf config.Config) error {

	// instance new gnb.
	gnb := &context.RanGnbContext{}

	// new gnb context.
	gnb.NewRanGnbContext(gnbId, conf.GNodeB.PlmnList.Mcc, conf.GNodeB.PlmnList.Mnc, conf.GNodeB.PlmnList.Tac, conf.GNodeB.SliceSupportList.St, conf.GNodeB.SliceSupportList.Sst)

	// send NGSetup request msg.
	gnbIdInBytes := gnb.GetGnbIdInBytes()
	err := interface_management.NgSetupRequest(connN2, gnbIdInBytes, 24, nameGNB)
	if logging.Check_error(err, "send NGSetupRequest Message") {
		return err
	}

	// receive NGSetupResponse Msg
	_, err = interface_management.NgSetupResponse(connN2)
	if logging.Check_error(err, "receive NGSetupResponse Message") {
		return err
	}

	// function worked fine.
	return nil
}
