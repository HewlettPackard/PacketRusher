/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package nas_transport

import (
	"encoding/binary"
	"fmt"
	nasIE "github.com/free5gc/nas/ie"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	codec "my5G-RANTester/lib/ngap"
)

func GetInitialUEMessage(ranUeNgapID int64, nasPdu []byte, guti5g *nasIE.MobileId5GS, gnb *context.GNBContext) ([]byte, error) {
	return BuildInitialUEMessage(ranUeNgapID, nasPdu, guti5g, gnb).MarshalBinary()
}
func BuildInitialUEMessage(ranUeNgapID int64, nasPdu []byte, guti5g *nasIE.MobileId5GS, gnb *context.GNBContext) *message.InitialUEMessage {
	value := &message.InitialUEMessage{
		RANUENGAPID:             &ie.RANUENGAPID{Value: ranUeNgapID},
		NASPDU:                  &ie.NASPDU{Value: nasPdu},
		UserLocationInformation: codec.UserLocation(gnb.GetPLMNIdentity(), gnb.GetNRCellIdentity(), gnb.GetTacInBytes()),
		RRCEstablishmentCause:   &ie.RRCEstablishmentCause{Value: ie.RRCEstablishmentCausePresentMoSignalling},
		UEContextRequest:        &ie.UEContextRequest{Value: ie.UEContextRequestPresentRequested},
	}
	if guti5g != nil {
		set := make([]byte, 2)
		binary.BigEndian.PutUint16(set, guti5g.AMFSetID<<6)
		tmsi := guti5g.TMSI5G[:]
		value.FiveGSTMSI = &ie.FiveGSTMSI{
			AMFSetID:   &ie.AMFSetID{Value: aper.BitString{Bytes: set, BitLength: 10}},
			AMFPointer: &ie.AMFPointer{Value: aper.BitString{Bytes: []byte{guti5g.AMFPointer << 2}, BitLength: 6}},
			FiveGTMSI:  &ie.FiveGTMSI{Value: tmsi},
		}
	}
	return value
}
func SendInitialUeMessage(registrationRequest []byte, ue *context.GNBUe, gnb *context.GNBContext) ([]byte, error) {
	value, err := GetInitialUEMessage(ue.GetRanUeId(), registrationRequest, ue.GetTMSI(), gnb)
	if err != nil {
		return nil, fmt.Errorf("build initial UE message for UE %d: %w", ue.GetRanUeId(), err)
	}
	return value, nil
}
