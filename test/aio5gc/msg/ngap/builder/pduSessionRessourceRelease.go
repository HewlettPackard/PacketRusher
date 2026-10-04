/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	codec "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc/context"
)

func PDUSessionResourceRelease(nas []byte, ue *context.UEContext, id int32) ([]byte, error) {
	msg, err := buildPDUSessionResourceRelease(nas, ue, id)
	if err != nil {
		return nil, err
	}
	return msg.MarshalBinary()
}
func buildPDUSessionResourceRelease(nas []byte, ue *context.UEContext, id int32) (*message.PDUSessionResourceReleaseCommand, error) {
	transfer := &ie.PDUSessionResourceReleaseCommandTransfer{Cause: &ie.Cause{Choice: &ie.CauseNas{Value: ie.CauseNasPresentNormalRelease}}}
	b, err := codec.Marshal(transfer)
	if err != nil {
		return nil, err
	}
	octets := aper.OctetString(b)
	return &message.PDUSessionResourceReleaseCommand{
		AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: ue.GetRanNgapId()}, NASPDU: &ie.NASPDU{Value: nas},
		PDUSessionResourceToReleaseListRelCmd: &ie.PDUSessionResourceToReleaseListRelCmd{List: []ie.PDUSessionResourceToReleaseItemRelCmd{{PDUSessionID: &ie.PDUSessionID{Value: int64(id)}, PDUSessionResourceReleaseCommandTransfer: &octets}}},
	}, nil
}
