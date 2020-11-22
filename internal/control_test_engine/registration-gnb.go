package control_test_engine

import (
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/context"
	"my5G-RANTester/internal/control_test_engine/ngap_control/interface_management"
)

func RegistrationGNB(connN2 *sctp.SCTPConn, gnbId string, nameGNB string, conf config.Config) (*context.RanGnbContext, error) {

	// instance new gnb.
	gnb := &context.RanGnbContext{}

	// new gnb context.
	gnb.NewRanGnbContext(gnbId, conf.GNodeB.PlmnList.Mcc, conf.GNodeB.PlmnList.Mnc, conf.GNodeB.PlmnList.Tac, conf.GNodeB.SliceSupportList.St, conf.GNodeB.SliceSupportList.Sst)

	// send NGSetup request msg.
	// gnbIdInBytes := gnb.GetGnbIdInBytes()
	log.WithFields(log.Fields{
		"protocol":    "ngap",
		"source":      "gNodeB",
		"destination": "AMF",
		"message":     "NGSetupRequest",
	}).Info("Sending message")
	err := interface_management.NgSetupRequest(connN2, gnb, 24, nameGNB)
	if err != nil {
		log.Error("Error sending message: NGSetupRequest")
		return nil, err
	}

	log.WithFields(log.Fields{
		"protocol":    "ngap",
		"source":      "AMF",
		"destination": "tgNodeB",
		"message":     "NgSetupResponse",
	}).Info("Sending message")
	// receive NGSetupResponse Msg
	_, err = interface_management.NgSetupResponse(connN2)
	if err != nil {
		log.Error("Error receiving message: NgSetupResponse")
		return nil, err
	}

	// function worked fine.
	return gnb, nil
}
