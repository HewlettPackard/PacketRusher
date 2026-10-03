/** SPDX-License-Identifier: Apache-2.0 */
package builder

import (
	"encoding/binary"
	"fmt"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	codec "my5G-RANTester/lib/ngap"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
	"net"
)

func PDUSessionResourceSetup(nas []byte, sm context.SmContext, ue *context.UEContext, session *context.SessionContext) ([]byte, error) {
	msg, err := buildPDUSessionResourceSetupRequest(ue, sm, nas, session.GetN3())
	if err != nil {
		return nil, err
	}
	return msg.MarshalBinary()
}
func buildPDUSessionResourceSetupRequest(ue *context.UEContext, sm context.SmContext, nas []byte, ip net.IP) (*message.PDUSessionResourceSetupRequest, error) {
	transfer, err := buildPDUSessionResourceSetuprequestTransfert(sm, ip)
	if err != nil {
		return nil, err
	}
	octets := aper.OctetString(transfer)
	ul, err := convert.BitRate(sm.GetSessionRule().AuthSessAmbr.Uplink)
	if err != nil {
		return nil, err
	}
	dl, err := convert.BitRate(sm.GetSessionRule().AuthSessAmbr.Downlink)
	if err != nil {
		return nil, err
	}
	return &message.PDUSessionResourceSetupRequest{
		AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: ue.GetRanNgapId()},
		UEAggregateMaximumBitRate:        &ie.UEAggregateMaximumBitRate{UEAggregateMaximumBitRateUL: &ie.BitRate{Value: ul}, UEAggregateMaximumBitRateDL: &ie.BitRate{Value: dl}},
		PDUSessionResourceSetupListSUReq: &ie.PDUSessionResourceSetupListSUReq{List: []ie.PDUSessionResourceSetupItemSUReq{{PDUSessionID: &ie.PDUSessionID{Value: int64(sm.GetPduSessionId())}, SNSSAI: convert.SNSSAIToNGAP(sm.GetSnnsai()), PDUSessionNASPDU: &ie.NASPDU{Value: nas}, PDUSessionResourceSetupRequestTransfer: &octets}}},
	}, nil
}
func buildPDUSessionResourceSetuprequestTransfert(sm context.SmContext, ip net.IP) ([]byte, error) {
	rule := sm.GetSessionRule()
	if rule == nil || rule.AuthSessAmbr == nil || rule.AuthDefQos == nil {
		return nil, fmt.Errorf("missing session QoS rule")
	}
	// The mock dataplane uses one deterministic TEID; it never carries user traffic.
	ul, err := convert.BitRate(rule.AuthSessAmbr.Uplink)
	if err != nil {
		return nil, err
	}
	dl, err := convert.BitRate(rule.AuthSessAmbr.Downlink)
	if err != nil {
		return nil, err
	}
	teid := make([]byte, 4)
	binary.BigEndian.PutUint32(teid, 1)
	transfer := &ie.PDUSessionResourceSetupRequestTransfer{ProtocolIEs: &ie.ProtocolIEContainerPDUSessionResourceSetupRequestTransferIEs{List: []ie.PDUSessionResourceSetupRequestTransferIEs{
		{PDUSessionAggregateMaximumBitRate: &ie.PDUSessionAggregateMaximumBitRate{PDUSessionAggregateMaximumBitRateUL: &ie.BitRate{Value: ul}, PDUSessionAggregateMaximumBitRateDL: &ie.BitRate{Value: dl}}},
		{ULNGUUPTNLInformation: &ie.UPTransportLayerInformation{Choice: &ie.GTPTunnel{TransportLayerAddress: &ie.TransportLayerAddress{Value: aper.BitString{Bytes: ip.To4(), BitLength: 32}}, GTPTEID: &ie.GTPTEID{Value: teid}}}},
		{PDUSessionType: &ie.PDUSessionType{Value: ie.PDUSessionTypePresentIpv4}},
		{QosFlowSetupRequestList: &ie.QosFlowSetupRequestList{List: []ie.QosFlowSetupRequestItem{{
			QosFlowIdentifier: &ie.QosFlowIdentifier{Value: int64(sm.GetDefQosQFI())},
			QosFlowLevelQosParameters: &ie.QosFlowLevelQosParameters{
				QosCharacteristics: &ie.QosCharacteristics{Choice: &ie.NonDynamic5QIDescriptor{FiveQI: &ie.FiveQI{Value: int64(rule.AuthDefQos.Var5qi)}}},
				AllocationAndRetentionPriority: &ie.AllocationAndRetentionPriority{
					PriorityLevelARP:        &ie.PriorityLevelARP{Value: int64(rule.AuthDefQos.Arp.PriorityLevel)},
					PreEmptionCapability:    &ie.PreEmptionCapability{Value: ie.PreEmptionCapabilityPresentShallNotTriggerPreEmption},
					PreEmptionVulnerability: &ie.PreEmptionVulnerability{Value: ie.PreEmptionVulnerabilityPresentNotPreEmptable},
				},
			},
		}}}},
	}}}
	return codec.Marshal(transfer)
}
