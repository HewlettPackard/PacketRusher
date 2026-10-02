package builder

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/test/aio5gc/context"
	"net"
	"testing"
)

func TestPDUSessionAcceptMarshalsProtectedIPv4QoSAndDNS(t *testing.T) {
	session := new(context.SessionContext)
	session.NewSessionContext()
	sm := context.NewSmContext(1)
	sm.SetPti(7)
	sm.SetPduSessionType(ie.PDUSessType_IPv4)
	sm.SetSnssai(models.Snssai{Sst: 1, Sd: "010203"})
	sm.SetPDUAddress(net.ParseIP("10.0.0.2"))
	sm.SetDefQosQFI(1)
	sm.SetSessionRule(session.GetSessionRules()[0])
	network, err := session.GetDataNetwork("internet")
	require.NoError(t, err)
	sm.SetDataNetwork(network)
	sm.ProtocolConfigurationOptions.DNSIPv4Request = true
	sm.ProtocolConfigurationOptions.DNSIPv6Request = true
	security := new(context.SecurityContext)
	security.SetIntegrityAlg(2)
	security.SetCipheringAlg(2)
	ue := new(context.UEContext)
	ue.SetSecurityContext(security)
	device := security.NASSecurityContext().Clone()
	device.Side = nas.UESide
	packet, err := PDUSessionEstablishmentAccept(ue, sm)
	require.NoError(t, err, "complete PDU accept must be encodable before sending")
	outer, err := nas.Parse(packet, device)
	require.NoError(t, err)
	downlink, ok := outer.(*nas.DLNASTransport)
	require.True(t, ok)
	require.Equal(t, uint8(1), downlink.PDUSessID.Value)
	inner, err := nas.Parse(downlink.PayloadCntr.Contents, nil)
	require.NoError(t, err)
	accept, ok := inner.(*nas.PDUSessEstAccept)
	require.True(t, ok)
	require.Equal(t, uint8(7), accept.PTI)
	require.Equal(t, []byte{10, 0, 0, 2}, accept.PDUAddr.IPv4)
	require.Equal(t, "internet", accept.DNN.Value)
	require.Equal(t, "010203", accept.SNSSAI.SD)
	require.Len(t, accept.AuthoQosRules.Rules, 1)
	rule := accept.AuthoQosRules.Rules[0]
	require.True(t, rule.IsDefaultDQR)
	require.Equal(t, uint8(1), rule.QFI)
	require.Len(t, rule.PktFilterList, 1)
	require.True(t, rule.PktFilterList[0].Contents.MatchAll)
	require.Equal(t, "8.8.8.8", accept.ExtendedProtCfgOpts.FromNw.DNSIPv4Addr.String())
	require.Equal(t, "2001:4860:4860::8888", accept.ExtendedProtCfgOpts.FromNw.DNSIPv6Addr.String())
}
