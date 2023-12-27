/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package pdu_session_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	"github.com/free5gc/ngap"

	"github.com/free5gc/ngap/ngapType"
)

func PDUSessionReleaseResponse(pduSessionIds []ngapType.PDUSessionID, ue *context.GNBUe) ([]byte, error) {

	message := buildPDUSessionReleaseResponse(ue.GetAmfUeId(), ue.GetRanUeId(), pduSessionIds)
	return ngap.Encoder(message)
}

func buildPDUSessionReleaseResponse(amfUeNgapID, ranUeNgapID int64, pduSessionIds []ngapType.PDUSessionID) (pdu ngapType.NGAPPDU) {

	pdu.Present = ngapType.NGAPPDUPresentSuccessfulOutcome
	pdu.SuccessfulOutcome = new(ngapType.SuccessfulOutcome)

	successfulOutcome := pdu.SuccessfulOutcome
	successfulOutcome.ProcedureCode.Value = ngapType.ProcedureCodePDUSessionResourceRelease
	successfulOutcome.Criticality.Value = ngapType.CriticalityPresentReject

	successfulOutcome.Value.Present = ngapType.SuccessfulOutcomePresentPDUSessionResourceReleaseResponse
	successfulOutcome.Value.PDUSessionResourceReleaseResponse = new(ngapType.PDUSessionResourceReleaseResponse)

	pDUSessionResourceReleaseResponse := successfulOutcome.Value.PDUSessionResourceReleaseResponse
	pDUSessionResourceReleaseResponseIEs := &pDUSessionResourceReleaseResponse.ProtocolIEs

	// AMF UE NGAP ID
	ie := ngapType.PDUSessionResourceReleaseResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.PDUSessionResourceSetupResponseIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	pDUSessionResourceReleaseResponseIEs.List = append(pDUSessionResourceReleaseResponseIEs.List, ie)

	// RAN UE NGAP ID
	ie = ngapType.PDUSessionResourceReleaseResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.PDUSessionResourceSetupResponseIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	pDUSessionResourceReleaseResponseIEs.List = append(pDUSessionResourceReleaseResponseIEs.List, ie)

	// PDU Session Resource Setup Response List
	ie = ngapType.PDUSessionResourceReleaseResponseIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDPDUSessionResourceReleasedListRelRes
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.PDUSessionResourceReleaseResponseIEsPresentPDUSessionResourceReleasedListRelRes
	ie.Value.PDUSessionResourceReleasedListRelRes = new(ngapType.PDUSessionResourceReleasedListRelRes)

	pDUSessionResourceReleasedListRelRes := ie.Value.PDUSessionResourceReleasedListRelRes

	for _, pduSessionId := range pduSessionIds {
		pDUSessionResourceReleasedItemRelRes := ngapType.PDUSessionResourceReleasedItemRelRes{}

		pDUSessionResourceReleasedItemRelRes.PDUSessionID = pduSessionId
		pDUSessionResourceReleasedItemRelRes.PDUSessionResourceReleaseResponseTransfer = []byte{00}

		pDUSessionResourceReleasedListRelRes.List = append(pDUSessionResourceReleasedListRelRes.List, pDUSessionResourceReleasedItemRelRes)
	}

	pDUSessionResourceReleaseResponseIEs.List = append(pDUSessionResourceReleaseResponseIEs.List, ie)

	return
}
