package builder

import (
	"encoding/hex"
	"github.com/free5gc/ngap/message"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/test/aio5gc/context"
	"testing"
)

// These packets were captured from the pre-migration builders at PacketRusher b61d814.
// Matching the old bytes guards protocol compatibility independently of the new decoder.
func TestNGSetupResponseWireCompatibility(t *testing.T) {
	plmn := models.PlmnId{Mcc: "208", Mnc: "93"}
	served := models.PlmnIdNid{Mcc: "208", Mnc: "93"}
	msg := BuilNGSetupResponse("amf.5gc.3gppnetwork.org", "196673", []models.Guami{{PlmnId: &served, AmfId: "196673"}}, []models.Nrf_NFMgmt_PlmnSnssai{{PlmnId: &plmn, SNssaiList: []models.ExtSnssai{{Sst: 1, Sd: "010203"}}}}, 100)
	b, err := msg.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, "20150040000004000100190b00616d662e3567632e336770706e6574776f726b2e6f726700600008000002f83919667300564001640050000b0002f83900001008010203", hex.EncodeToString(b))
	decoded, err := message.Parse(b)
	require.NoError(t, err)
	require.IsType(t, &message.NGSetupResponse{}, decoded)
}
func TestDownlinkNASWireCompatibility(t *testing.T) {
	ue := new(context.UEContext)
	ue.SetRanNgapId(12)
	ue.SetAmfNgapId(42)
	b, err := DownlinkNASTransport([]byte{0x7e, 0, 0x57}, ue)
	require.NoError(t, err)
	require.Equal(t, "00044017000003000a0002002a00550002000c00260004037e0057", hex.EncodeToString(b))
}
