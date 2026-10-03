/** SPDX-License-Identifier: Apache-2.0 */
package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/msg/nas/codec"
)

func PDUSessionEstablishmentAccept(ue *context.UEContext, sm *context.SmContext) ([]byte, error) {
	accept, err := buildSessionEstablishmentAccept(ue, sm)
	if err != nil {
		return nil, err
	}
	b, err := accept.MarshalBinary()
	if err != nil {
		return nil, err
	}
	msg, err := buildDLNASTransport(ue, b, uint8(sm.GetPduSessionId()))
	if err != nil {
		return nil, err
	}
	return codec.Encode(ue, msg, nas.SecHdrTypeIntegrityProtectedAndCiphered)
}
func buildDLNASTransport(ue *context.UEContext, payload []byte, id uint8) (*nas.DLNASTransport, error) {
	return &nas.DLNASTransport{
		PayloadCntrType: &ie.PayloadCntrType{Value: ie.PayloadCntrType_N1SMInfo},
		PayloadCntr:     &ie.PayloadCntr{Pct: ie.PayloadCntrType_N1SMInfo, Contents: payload},
		PDUSessID:       &ie.PDUSessId2{Value: id},
	}, nil
}
func buildSessionEstablishmentAccept(ue *context.UEContext, sm *context.SmContext) (*nas.PDUSessEstAccept, error) {
	rule := sm.GetSessionRule()
	ambr := new(ie.SessAMBR)
	if err := ambr.Set(rule.AuthSessAmbr.Uplink, rule.AuthSessAmbr.Downlink); err != nil {
		return nil, err
	}
	address, _ := sm.PDUAddressToNAS()
	snssai := sm.GetSnnsai()
	msg := &nas.PDUSessEstAccept{
		PDUSessId: uint8(sm.GetPduSessionId()), PTI: sm.GetPti(),
		SelectedPDUSessType: &ie.PDUSessType{Value: sm.GetPduSessionType()}, SelectedSSCMode: &ie.SSCMode{Mode: 1},
		AuthoQosRules: &ie.QosRules{Rules: []ie.QosRule{{RuleId: 1, IsDefaultDQR: true, OpCode: ie.OpCode_CreateNewQosRule, Precedence: 255, QFI: sm.GetDefQosQFI(), PktFilterList: []ie.PacketFilter{{Id: 1, Dir: 3, Contents: ie.PacketFilterContents{MatchAll: true, RemoteAddr: "any", LocalAddr: "any"}}}}}},
		SessAMBR:      ambr, PDUAddr: &ie.PDUAddr{IPv4: address[:4]}, SNSSAI: &ie.SNSSAI{SST: uint8(snssai.Sst), SD: snssai.Sd},
		AuthoQosFlowDescs: &ie.QosFlowDescs{Descs: []ie.QosFlowDesc{{QFI: sm.GetDefQosQFI(), OpCode: ie.QFD_Create, EBit: 1, FiveQI: uint8(rule.AuthDefQos.Var5qi)}}},
		DNN:               &ie.DNN{Value: sm.GetDataNetwork().Dnn},
	}
	options := sm.ProtocolConfigurationOptions
	if options.DNSIPv4Request || options.DNSIPv6Request {
		from := new(ie.ExtCfgOptFromNw)
		if options.DNSIPv4Request {
			from.DNSIPv4Addr = sm.GetDataNetwork().Dns.IPv4Addr
		}
		if options.DNSIPv6Request {
			from.DNSIPv6Addr = sm.GetDataNetwork().Dns.IPv6Addr
		}
		msg.ExtendedProtCfgOpts = &ie.ExtendedProtCfgOpts{FromNw: from}
	}
	return msg, nil
}
