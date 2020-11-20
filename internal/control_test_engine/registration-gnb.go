package control_test_engine

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/context"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
	"my5G-RANTester/internal/logging"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId string, nameGNB string, conf config.Config) (*context.RanGnbContext, error) {

	// instance new gnb.
	gnb := &context.RanGnbContext{}

	// new gnb context.
	gnb.NewRanGnbContext(gnbId, conf.GNodeB.PlmnList.Mcc, conf.GNodeB.PlmnList.Mnc, conf.GNodeB.PlmnList.Tac, conf.GNodeB.SliceSupportList.St, conf.GNodeB.SliceSupportList.Sst)

	// send NGSetup request msg.
	// gnbIdInBytes := gnb.GetGnbIdInBytes()
	err := interface_management.NgSetupRequest(connN2, gnb, 24, nameGNB)
	msg := fmt.Sprintf("[NGAP][GNB ID %s]SEND NG SETUP REQUEST", gnbId)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	// receive NGSetupResponse Msg
	_, err = interface_management.NgSetupResponse(connN2)
	msg = fmt.Sprintf("[NGAP][GNB ID %s]RECEIVE NG SETUP RESPONSE", gnbId)
	if logging.Check_error(err, msg) {
		return nil, err
	}

	fmt.Println("GNB AUTHENTICATION FINISHED")

	// function worked fine.
	return gnb, nil
}
