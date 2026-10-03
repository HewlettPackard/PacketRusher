/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ue_context_management

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"

	ngap "github.com/free5gc/ngap/message"

	ngapType "github.com/free5gc/ngap/ie"
)

/*
func initialContextSetupResponse(connN2 *sctp.SCTPConn, amfUeNgapID int64, ranUeNgapID int64, supi string) error {

	sendMsg, err := InitialContextSetupResponse(amfUeNgapID, ranUeNgapID)
	if err != nil {
		return fmt.Errorf("Error getting %s ue ngap Initial Context Setup Response Msg", supi)
	}
	_, err = connN2.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("Error sending %s ue Initial Context Setup Response Msg", supi)
	}

	return nil
}
*/

func UeContextReleaseComplete(ue *context.GNBUe) ([]byte, error) {
	message := BuildUeContextReleaseComplete(ue.GetAmfUeId(), ue.GetRanUeId())

	return message.MarshalBinary()
}

func BuildUeContextReleaseComplete(amfUeNgapID, ranUeNgapID int64) (pdu *ngap.UEContextReleaseComplete) {
	pdu = &ngap.UEContextReleaseComplete{}

	// AMF UE NGAP ID
	ie := ngapType.UEContextReleaseCompleteIEs{}

	ie.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.AMFUENGAPID
	aMFUENGAPID.Value = amfUeNgapID

	pdu.AMFUENGAPID = ie.AMFUENGAPID

	// RAN UE NGAP ID
	ie = ngapType.UEContextReleaseCompleteIEs{}

	ie.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.RANUENGAPID
	rANUENGAPID.Value = ranUeNgapID

	pdu.RANUENGAPID = ie.RANUENGAPID

	return
}
