package control_test_engine

import (
	"fmt"
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
		"protocol":    "NGAP",
		"source":      fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"destination": "AMF",
		"message":     "NG SETUP REQUEST",
	}).Info("Sending message")
	err := interface_management.NgSetupRequest(connN2, gnb, 24, nameGNB)
	if err != nil {
		log.Error("Error sending message: NGSetupRequest")
		return nil, err
	}

	log.WithFields(log.Fields{
		"protocol":    "NGAP",
		"source":      "AMF",
		"destination": fmt.Sprintf("GNB[ID:%s]", gnb.GetGnbId()),
		"message":     "NG SETUP RESPONSE",
	}).Info("Receive message")
	// receive NGSetupResponse Msg
	_, err = interface_management.NgSetupResponse(connN2)
	if err != nil {
		log.Error("Error receiving message: NgSetupResponse")
		return nil, err
	}

	//fmt.Println("[GNB] GNB AUTHENTICATION FINISHED")

	// function worked fine.
	return gnb, nil
}
