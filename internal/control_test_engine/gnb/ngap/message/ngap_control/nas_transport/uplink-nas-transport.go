/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas_transport

import (
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	codec "my5G-RANTester/lib/ngap"
)

func buildUplinkNasTransport(amfUeNgapID, ranUeNgapID int64, nasPdu []byte, gnb *context.GNBContext) *message.UplinkNASTransport {
	return &message.UplinkNASTransport{
		AMFUENGAPID:             &ie.AMFUENGAPID{Value: amfUeNgapID},
		RANUENGAPID:             &ie.RANUENGAPID{Value: ranUeNgapID},
		NASPDU:                  &ie.NASPDU{Value: nasPdu},
		UserLocationInformation: codec.UserLocation(gnb.GetPLMNIdentity(), gnb.GetNRCellIdentity(), gnb.GetTacInBytes()),
	}
}
func SendUplinkNasTransport(payload []byte, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	return buildUplinkNasTransport(ue.GetAmfUeId(), ue.GetRanUeId(), payload, gnb).MarshalBinary()
}
