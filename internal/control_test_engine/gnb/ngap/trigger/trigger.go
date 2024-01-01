/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package trigger

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	ueSender "my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/interface_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_context_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_mobility_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"

	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func SendPduSessionResourceSetupResponse(pduSessions []*context.GnbPDUSession, ue *context.GNBUe, gnb *context.GNBContext) {
	log.Info("[GNB] Initiating PDU Session Resource Setup Response")

	// send PDU Session Resource Setup Response.
	ngapMsg, err := pdu_session_management.PDUSessionResourceSetupResponse(pduSessions, ue, gnb)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending PDU Session Resource Setup Response: ", err)
	}

	ue.SetStateReady()

	// Send PDU Session Resource Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][AMF] Error sending PDU Session Resource Setup Response.: ", err)
	}
}

func SendPduSessionReleaseResponse(pduSessionIds []ngapType.PDUSessionID, ue *context.GNBUe) {
	log.Info("[GNB] Initiating PDU Session Release Response")

	if len(pduSessionIds) == 0 {
		log.Fatal("[GNB][NGAP] Trying to send a PDU Session Release Reponse for no PDU Session")
	}

	ngapMsg, err := pdu_session_management.PDUSessionReleaseResponse(pduSessionIds, ue)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending PDU Session Release Response.: ", err)
	}

	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending PDU Session Release Response.: ", err)
	}
}

func SendInitialContextSetupResponse(ue *context.GNBUe, gnb *context.GNBContext) {
	log.Info("[GNB] Initiating Initial Context Setup Response")

	// send Initial Context Setup Response.
	ngapMsg, err := ue_context_management.InitialContextSetupResponse(ue, gnb)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending Initial Context Setup Response: ", err)
	}

	// Send Initial Context Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][AMF] Error sending Initial Context Setup Response: ", err)
	}
}

func SendUeContextReleaseRequest(ue *context.GNBUe) {
	log.Info("[GNB] Initiating UE Context Release Request")

	// send UE Context Release Complete
	ngapMsg, err := ue_context_management.UeContextReleaseRequest(ue)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending UE Context Release Request: ", err)
	}

	// Send UE Context Release Complete
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][AMF] Error sending UE Context Release Request: ", err)
	}
}

func SendUeContextReleaseComplete(ue *context.GNBUe) {
	log.Info("[GNB] Initiating UE Context Complete")

	// send UE Context Release Complete
	ngapMsg, err := ue_context_management.UeContextReleaseComplete(ue)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending UE Context Complete: ", err)
	}

	// Send UE Context Release Complete
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][AMF] Error sending UE Context Complete: ", err)
	}
}

func SendAmfConfigurationUpdateAcknowledge(amf *context.GNBAmf) {
	log.Info("[GNB] Initiating AMF Configuration Update Acknowledge")

	// send AMF Configure Update Acknowledge
	ngapMsg, err := interface_management.AmfConfigurationUpdateAcknowledge()
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending AMF Configuration Update Acknowledge: ", err)
	}

	// Send AMF Configure Update Acknowledge
	conn := amf.GetSCTPConn()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending AMF Configuration Update Acknowledge: ", err)
	}
}

func SendNgSetupRequest(gnb *context.GNBContext, amf *context.GNBAmf) {
	log.Info("[GNB] Initiating NG Setup Request")

	// send NG setup response.
	ngapMsg, err := interface_management.NGSetupRequest(gnb, "PacketRusher")
	if err != nil {
		log.Info("[GNB][NGAP] Error sending NG Setup Request: ", err)
	}

	conn := amf.GetSCTPConn()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Info("[GNB][AMF] Error sending NG Setup Request: ", err)
	}

}

func SendPathSwitchRequest(gnb *context.GNBContext, ue *context.GNBUe) {
	log.Info("[GNB] Initiating Path Switch Request")

	// send NG setup response.
	ngapMsg, err := ue_mobility_management.PathSwitchRequest(gnb, ue)
	if err != nil {
		log.Info("[GNB][NGAP] Error sending Path Switch Request ", err)
	}

	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending Path Switch Request: ", err)
	}
}

func SendHandoverRequestAcknowledge(gnb *context.GNBContext, ue *context.GNBUe) {
	log.Info("[GNB] Initiating Handover Request Acknowledge")

	// send NG setup response.
	ngapMsg, err := ue_mobility_management.HandoverRequestAcknowledge(gnb, ue)
	if err != nil {
		log.Info("[GNB][NGAP] Error sending Handover Request Acknowledge: ", err)
	}

	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending Handover Request Acknowledge: ", err)
	}
}

func SendHandoverNotify(gnb *context.GNBContext, ue *context.GNBUe) {
	log.Info("[GNB] Initiating Handover Notify")

	// send NG setup response.
	ngapMsg, err := ue_mobility_management.HandoverNotify(gnb, ue)
	if err != nil {
		log.Info("[GNB][NGAP] Error sending Handover Notify: ", err)
	}

	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending Handover Notify: ", err)
	}
}

func TriggerXnHandover(oldGnb *context.GNBContext, newGnb *context.GNBContext, prUeId int64) {
	log.Info("[GNB] Initiating Xn UE Handover")

	gnbUeContext, err := oldGnb.GetGnbUeByPrUeId(prUeId)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error getting UE from PR UE ID: ", err)
	}

	newGnbRx := make(chan context.UEMessage, 1)
	newGnbTx := make(chan context.UEMessage, 1)
	newGnb.GetInboundChannel() <- context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, PrUeId: gnbUeContext.GetPrUeId(), UEContext: gnbUeContext, IsHandover: true}

	msg := context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, GNBInboundChannel: newGnb.GetInboundChannel()}

	ueSender.SendMessageToUe(gnbUeContext, msg)
}

func TriggerNgapHandover(oldGnb *context.GNBContext, newGnb *context.GNBContext, prUeId int64) {
	log.Info("[GNB] Initiating NGAP UE Handover")

	gnbUeContext, err := oldGnb.GetGnbUeByPrUeId(prUeId)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error getting UE from PR UE ID: ", err)
	}

	gnbUeContext.SetHandoverGnodeB(newGnb)

	// send NG setup response.
	ngapMsg, err := ue_mobility_management.HandoverRequired(oldGnb, newGnb, gnbUeContext)
	if err != nil {
		log.Info("[GNB][NGAP] Error sending Handover Required ", err)
	}

	conn := gnbUeContext.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Fatal("[GNB][NGAP] Error sending Handover Required: ", err)
	}
}
