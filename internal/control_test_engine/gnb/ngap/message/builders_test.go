package message_test

import (
	"encoding/binary"
	"net/netip"
	"reflect"
	"testing"

	nasIE "github.com/free5gc/nas/ie"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	interfaces "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/interface_management"
	nas "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/nas_transport"
	pdu "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/pdu_session_management"
	uecontext "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_context_management"
	mobility "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_mobility_management"
	codec "my5G-RANTester/lib/ngap"
)

func fixture(t *testing.T, id string) (*context.GNBContext, *context.GNBUe, *context.GnbPDUSession) {
	t.Helper()
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext(id, "208", "93", "000001", "01", "000001", netip.MustParseAddrPort("127.0.0.1:9487"), netip.MustParseAddrPort("127.0.0.1:2152"))
	gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412")).SetStateActive()
	ue, err := gnb.NewGnBUe(make(chan context.UEMessage, 10), make(chan context.UEMessage, 10), 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	bits := aper.BitString{Bytes: []byte{0, 0}, BitLength: 16}
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"000001"}, &ie.UESecurityCapabilities{
		NRencryptionAlgorithms:             &ie.NRencryptionAlgorithms{Value: bits},
		NRintegrityProtectionAlgorithms:    &ie.NRintegrityProtectionAlgorithms{Value: bits},
		EUTRAencryptionAlgorithms:          &ie.EUTRAencryptionAlgorithms{Value: bits},
		EUTRAintegrityProtectionAlgorithms: &ie.EUTRAintegrityProtectionAlgorithms{Value: bits},
	})
	ue.SetAmfUeId(42)
	session, err := ue.CreatePduSession(1, "10.0.0.1", "01", "000001", 0, 9, 1, 9, 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	return gnb, ue, session
}

// Parsing real encoded bytes exercises the upstream mandatory-field checks,
// nested pointer initialization, CHOICE selection and transparent transfer codecs.
func TestNativeBuildersWireRoundTrip(t *testing.T) {
	gnb, ue, session := fixture(t, "000001")
	target, _, _ := fixture(t, "000002")
	cases := []struct {
		name  string
		build func() ([]byte, error)
		want  message.Message
	}{
		{"NG setup", func() ([]byte, error) { return interfaces.NGSetupRequest(gnb, "test") }, &message.NGSetupRequest{}},
		{"AMF configuration acknowledge", interfaces.AmfConfigurationUpdateAcknowledge, &message.AMFConfigurationUpdateAcknowledge{}},
		{"initial UE", func() ([]byte, error) {
			return nas.GetInitialUEMessage(ue.GetRanUeId(), []byte{0x7e, 0, 0x41}, nil, gnb)
		}, &message.InitialUEMessage{}},
		{"uplink NAS", func() ([]byte, error) { return nas.SendUplinkNasTransport([]byte{0x7e, 0, 0x41}, ue, gnb) }, &message.UplinkNASTransport{}},
		{"initial context response", func() ([]byte, error) { return uecontext.InitialContextSetupResponse(ue, gnb) }, &message.InitialContextSetupResponse{}},
		{"PDU setup response", func() ([]byte, error) {
			return pdu.PDUSessionResourceSetupResponse([]*context.GnbPDUSession{session}, ue, gnb)
		}, &message.PDUSessionResourceSetupResponse{}},
		{"PDU release response", func() ([]byte, error) { return pdu.PDUSessionReleaseResponse([]*ie.PDUSessionID{{Value: 1}}, ue) }, &message.PDUSessionResourceReleaseResponse{}},
		{"UE release request", func() ([]byte, error) { return uecontext.UeContextReleaseRequest(ue) }, &message.UEContextReleaseRequest{}},
		{"UE release complete", func() ([]byte, error) { return uecontext.UeContextReleaseComplete(ue) }, &message.UEContextReleaseComplete{}},
		{"handover required", func() ([]byte, error) { return mobility.HandoverRequired(gnb, target, ue) }, &message.HandoverRequired{}},
		{"handover acknowledge", func() ([]byte, error) { return mobility.HandoverRequestAcknowledge(gnb, ue) }, &message.HandoverRequestAcknowledge{}},
		{"handover notify", func() ([]byte, error) { return mobility.HandoverNotify(gnb, ue) }, &message.HandoverNotify{}},
		{"path switch", func() ([]byte, error) { return mobility.PathSwitchRequest(gnb, ue) }, &message.PathSwitchRequest{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := tc.build()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := message.Parse(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(decoded) != reflect.TypeOf(tc.want) {
				t.Fatalf("got %T, want %T", decoded, tc.want)
			}
			switch value := decoded.(type) {
			case *message.PDUSessionResourceSetupResponse:
				transfer := &ie.PDUSessionResourceSetupResponseTransfer{}
				if err := codec.Unmarshal(value.PDUSessionResourceSetupListSURes.List[0].PDUSessionResourceSetupResponseTransfer, transfer); err != nil {
					t.Fatal(err)
				}
				tunnel, err := codec.Tunnel(transfer.DLQosFlowPerTNLInformation.UPTransportLayerInformation)
				if err != nil {
					t.Fatal(err)
				}
				if binary.BigEndian.Uint32(tunnel.GTPTEID.Value) != 200 {
					t.Fatal("downlink TEID changed during encoding")
				}
			case *message.HandoverRequired:
				transfer := &ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
				if err := codec.Unmarshal(value.SourceToTargetTransparentContainer.Value, transfer); err != nil {
					t.Fatal(err)
				}
				if transfer.IndexToRFSP.Value != 1 {
					t.Fatal("UE identifier changed in transparent container")
				}
			}
		})
	}
}

func TestInitialUEGUTIBitFieldsRoundTrip(t *testing.T) {
	gnb, ue, _ := fixture(t, "000001")
	guti := &nasIE.MobileId5GS{AMFSetID: 0x2a5, AMFPointer: 0x35, TMSI5G: [4]byte{1, 2, 3, 4}}
	encoded, err := nas.GetInitialUEMessage(ue.GetRanUeId(), []byte{0x7e, 0, 0x41}, guti, gnb)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := message.Parse(encoded)
	if err != nil {
		t.Fatal(err)
	}
	tmsi := decoded.(*message.InitialUEMessage).FiveGSTMSI
	if binary.BigEndian.Uint16(tmsi.AMFSetID.Value.Bytes)>>6 != guti.AMFSetID || tmsi.AMFPointer.Value.Bytes[0]>>2 != guti.AMFPointer {
		t.Fatal("GUTI bit fields changed during encoding")
	}
	if string(tmsi.FiveGTMSI.Value) != string(guti.TMSI5G[:]) {
		t.Fatal("TMSI changed during encoding")
	}
}

func TestHandoverTransparentIdentityBeyondRFSPRange(t *testing.T) {
	source, ue, _ := fixture(t, "000001")
	target, _, _ := fixture(t, "000002")
	for _, id := range []int64{257, 100000} {
		wire := mobility.GetSourceToTargetTransparentTransfer(source, target, ue.GetPduSessions(), id)
		container := &ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
		if err := codec.Unmarshal(wire, container); err != nil {
			t.Fatal(err)
		}
		got, err := codec.VirtualUEID(container)
		if err != nil || got != id {
			t.Fatalf("handover identity %d: %d %v", id, got, err)
		}
		if container.IndexToRFSP != nil {
			t.Fatal("unbounded UE identity used as constrained RFSP index")
		}
	}
}
