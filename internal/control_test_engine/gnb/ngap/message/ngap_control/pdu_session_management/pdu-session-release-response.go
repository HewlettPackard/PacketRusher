/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package pdu_session_management

import (
ngapConvert "my5G-RANTester/lib/ngap"
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	ngap "github.com/free5gc/ngap/message"

	ngapType "github.com/free5gc/ngap/ie"
)

func PDUSessionReleaseResponse(pduSessionIds []*ngapType.PDUSessionID, ue *context.GNBUe) ([]byte, error) {

	message := buildPDUSessionReleaseResponse(ue.GetAmfUeId(), ue.GetRanUeId(), pduSessionIds)
	return message.MarshalBinary()
}

func buildPDUSessionReleaseResponse(amfUeNgapID, ranUeNgapID int64, pduSessionIds []*ngapType.PDUSessionID) (pdu *ngap.PDUSessionResourceReleaseResponse) {
	pdu = &ngap.PDUSessionResourceReleaseResponse{}

	// AMF UE NGAP ID
	ie := ngapType.PDUSessionResourceReleaseResponseIEs{}

	ie.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	pdu.AMFUENGAPID = ie.AMFUENGAPID

	// RAN UE NGAP ID
	ie = ngapType.PDUSessionResourceReleaseResponseIEs{}

	ie.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	pdu.RANUENGAPID = ie.RANUENGAPID

	// PDU Session Resource Setup Response List
	ie = ngapType.PDUSessionResourceReleaseResponseIEs{}

	ie.PDUSessionResourceReleasedListRelRes = new(ngapType.PDUSessionResourceReleasedListRelRes)

	pDUSessionResourceReleasedListRelRes := ie.PDUSessionResourceReleasedListRelRes

	for _, pduSessionId := range pduSessionIds {
		pDUSessionResourceReleasedItemRelRes := ngapType.PDUSessionResourceReleasedItemRelRes{}

		pDUSessionResourceReleasedItemRelRes.PDUSessionID = pduSessionId
		pDUSessionResourceReleasedItemRelRes.PDUSessionResourceReleaseResponseTransfer = ngapConvert.Octets([]byte{0})

		pDUSessionResourceReleasedListRelRes.List = append(pDUSessionResourceReleasedListRelRes.List, pDUSessionResourceReleasedItemRelRes)
	}

	pdu.PDUSessionResourceReleasedListRelRes = ie.PDUSessionResourceReleasedListRelRes

	return
}
