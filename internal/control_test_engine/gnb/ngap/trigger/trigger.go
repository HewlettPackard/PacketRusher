/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package trigger

import (
	stdContext "context"
	"fmt"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	ueSender "my5G-RANTester/internal/control_test_engine/gnb/nas/message/sender"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/interface_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_context_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_mobility_management"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/message/sender"

	ngapType "github.com/free5gc/ngap/ie"
	log "github.com/sirupsen/logrus"
)

func SendPduSessionResourceSetupResponse(pduSessions []*context.GnbPDUSession, ue *context.GNBUe, gnb *context.GNBContext) {
	log.Info("[GNB] Initiating PDU Session Resource Setup Response")

	// send PDU Session Resource Setup Response.
	ngapMsg, err := pdu_session_management.PDUSessionResourceSetupResponse(pduSessions, ue, gnb)
	if err != nil {
		// A single UE's message must not end the process: one gNB carries hundreds of
		// others, and killing it discards every working session along with the bad one.
		log.Error("[GNB][NGAP] Error building PDU Session Resource Setup Response: ", err)
		return
	}

	ue.SetStateReady()

	// Send PDU Session Resource Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Error("[GNB][AMF] Error sending PDU Session Resource Setup Response: ", err)
	}
}

func SendPduSessionReleaseResponse(pduSessionIds []*ngapType.PDUSessionID, ue *context.GNBUe) {
	log.Info("[GNB] Initiating PDU Session Release Response")

	if len(pduSessionIds) == 0 {
		log.Error("[GNB][NGAP] Trying to send a PDU Session Release Reponse for no PDU Session")
		return
	}

	ngapMsg, err := pdu_session_management.PDUSessionReleaseResponse(pduSessionIds, ue)
	if err != nil {
		log.Error("[GNB][NGAP] Error sending PDU Session Release Response.: ", err)
		return
	}

	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Error("[GNB][NGAP] Error sending PDU Session Release Response.: ", err)
	}
}

func SendInitialContextSetupResponse(ue *context.GNBUe, gnb *context.GNBContext) {
	log.Info("[GNB] Initiating Initial Context Setup Response")

	// send Initial Context Setup Response.
	ngapMsg, err := ue_context_management.InitialContextSetupResponse(ue, gnb)
	if err != nil {
		log.Error("[GNB][NGAP] Error sending Initial Context Setup Response: ", err)
		return
	}

	// Send Initial Context Setup Response.
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Error("[GNB][AMF] Error sending Initial Context Setup Response: ", err)
	}
}

func SendUeContextReleaseRequest(ue *context.GNBUe) {
	log.Info("[GNB] Initiating UE Context Release Request")

	// send UE Context Release Request
	ngapMsg, err := ue_context_management.UeContextReleaseRequest(ue)
	if err != nil {
		log.Error("[GNB][NGAP] Error sending UE Context Release Request: ", err)
		return
	}

	// Only once there is a request to send: an Error Indication for this UE is
	// treated as the answer to it. Set before the send, so an answer that
	// arrives before the send returns still finds it.
	alreadyRequested := ue.GetReleaseRequested()
	ue.SetReleaseRequested(true)

	// Send UE Context Release Request
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Error("[GNB][AMF] Error sending UE Context Release Request: ", err)
		// Nothing was sent, so no Error Indication can answer it. An earlier
		// request that was sent is still pending, so leave its flag alone.
		if !alreadyRequested {
			ue.SetReleaseRequested(false)
		}
	}
}

func SendUeContextReleaseComplete(ue *context.GNBUe) {
	log.Info("[GNB] Initiating UE Context Complete")

	// send UE Context Release Complete
	ngapMsg, err := ue_context_management.UeContextReleaseComplete(ue)
	if err != nil {
		log.Error("[GNB][NGAP] Error sending UE Context Complete: ", err)
		return
	}

	// Send UE Context Release Complete
	conn := ue.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Error("[GNB][AMF] Error sending UE Context Complete: ", err)
	}
}

func SendAmfConfigurationUpdateAcknowledge(amf *context.GNBAmf) {
	log.Info("[GNB] Initiating AMF Configuration Update Acknowledge")

	// send AMF Configure Update Acknowledge
	ngapMsg, err := interface_management.AmfConfigurationUpdateAcknowledge()
	if err != nil {
		log.Warn("[GNB][NGAP] Error sending AMF Configuration Update Acknowledge: ", err)
	}

	// Send AMF Configure Update Acknowledge
	conn := amf.GetSCTPConn()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		log.Warn("[GNB][NGAP] Error sending AMF Configuration Update Acknowledge: ", err)
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
		log.Error("[GNB][NGAP] Error sending Path Switch Request: ", err)
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
		log.Error("[GNB][NGAP] Error sending Handover Request Acknowledge: ", err)
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
		log.Error("[GNB][NGAP] Error sending Handover Notify: ", err)
	}
}

func TriggerXnHandover(oldGnb *context.GNBContext, newGnb *context.GNBContext, prUeId int64) {
	log.Info("[GNB] Initiating Xn UE Handover")

	gnbUeContext, err := oldGnb.GetGnbUeByPrUeId(prUeId)
	if err != nil {
		log.Error("[GNB][NGAP] Error getting UE from PR UE ID: ", err)
		return
	}

	newGnbRx := make(chan context.UEMessage, 1)
	newGnbTx := make(chan context.UEMessage, 1)
	connectionLost := make(chan struct{})
	newGnb.GetInboundChannel() <- context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, ConnectionLost: connectionLost, PrUeId: gnbUeContext.GetPrUeId(), UEContext: gnbUeContext, IsHandover: true}

	msg := context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, ConnectionLost: connectionLost, GNBInboundChannel: newGnb.GetInboundChannel()}

	ueSender.SendMessageToUe(gnbUeContext, msg)
}

func TriggerNgapHandover(oldGnb *context.GNBContext, newGnb *context.GNBContext, prUeId int64) {
	if err := StartNgapHandover(oldGnb, newGnb, prUeId); err != nil {
		log.Error("[GNB][NGAP] Error initiating handover: ", err)
	}
}

// StartNgapHandover reports a build/send failure to runtime scenario callers.
func StartNgapHandover(oldGnb *context.GNBContext, newGnb *context.GNBContext, prUeId int64) error {
	log.Info("[GNB] Initiating NGAP UE Handover")

	gnbUeContext, err := oldGnb.GetGnbUeByPrUeId(prUeId)
	if err != nil {
		log.Error("[GNB][NGAP] Error getting UE from PR UE ID: ", err)
		return err
	}

	gnbUeContext.LockProcessing()
	gnbUeContext.SetHandoverGnodeB(newGnb)

	// send NG setup response.
	ngapMsg, err := ue_mobility_management.HandoverRequired(oldGnb, newGnb, gnbUeContext)
	gnbUeContext.UnlockProcessing()
	if err != nil {
		gnbUeContext.SetHandoverGnodeB(nil)
		return err
	}

	conn := gnbUeContext.GetSCTP()
	err = sender.SendToAmF(ngapMsg, conn)
	if err != nil {
		gnbUeContext.SetHandoverGnodeB(nil)
	}
	return err
}

// PrepareXnHandover returns the connection change to the UE event loop. Sending
// it to that loop's own downlink queue would deadlock when the queue is full.
func PrepareXnHandover(ctx stdContext.Context, oldGnb, newGnb *context.GNBContext, prUeId int64) (context.UEMessage, error) {
	if oldGnb == nil || newGnb == nil || oldGnb == newGnb {
		return context.UEMessage{}, fmt.Errorf("invalid handover source or target")
	}
	gu, err := oldGnb.GetGnbUeByPrUeId(prUeId)
	if err != nil {
		return context.UEMessage{}, err
	}
	if !newGnb.NGSetupReady() {
		return context.UEMessage{}, fmt.Errorf("target gNB has not completed NG Setup")
	}
	rx, tx := make(chan context.UEMessage, 10), make(chan context.UEMessage, 10)
	lost := make(chan struct{})
	message := context.UEMessage{GNBRx: rx, GNBTx: tx, ConnectionLost: lost, PrUeId: prUeId, UEContext: gu, IsHandover: true}
	if err := ctx.Err(); err != nil {
		return context.UEMessage{}, err
	}
	if err := newGnb.QueueHandover(message); err != nil {
		return context.UEMessage{}, err
	}
	return context.UEMessage{GNBRx: rx, GNBTx: tx, ConnectionLost: lost, GNBInboundChannel: newGnb.GetInboundChannel()}, nil
}
