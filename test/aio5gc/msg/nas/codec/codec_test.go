package codec

import (
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/test/aio5gc/context"
	"testing"
)

func newSecurityUE() *context.UEContext {
	ue := new(context.UEContext)
	security := new(context.SecurityContext)
	security.SetCipheringAlg(2)
	security.SetIntegrityAlg(2)
	ue.SetSecurityContext(security)
	return ue
}
func TestProtectedNASCountsAndIntegrity(t *testing.T) {
	ue := newSecurityUE()
	core := ue.GetSecurityContext().NASSecurityContext()
	device := core.Clone()
	device.Side = nas.UESide
	packet, err := nas.Marshal(&nas.RegComplete{}, device, nas.SecHdrTypeIntegrityProtectedAndCiphered)
	require.NoError(t, err)
	parsed, verified, err := Decode(ue, packet, false)
	require.NoError(t, err)
	require.True(t, verified)
	require.IsType(t, &nas.RegComplete{}, parsed)
	before := core.UplinkCount.Count
	tampered := append([]byte(nil), packet...)
	tampered[2] ^= 1
	_, verified, err = Decode(ue, tampered, false)
	require.Error(t, err)
	require.False(t, verified)
	require.Equal(t, before, core.UplinkCount.Count, "invalid MAC must not advance receive counter")
	downlink, err := Encode(ue, &nas.IdReq{IdType: &ie.IdType5GS{IdType: ie.IdType_5GS_SUCI}}, nas.SecHdrTypeIntegrityProtectedAndCiphered)
	require.NoError(t, err)
	require.Equal(t, uint32(1), core.DownlinkCount.Count)
	require.Equal(t, before, core.UplinkCount.Count, "downlink transmit must leave uplink counter alone")
	parsed, err = nas.Parse(downlink, device)
	require.NoError(t, err)
	require.IsType(t, &nas.IdReq{}, parsed)
}
func TestPlainDecoderRejectsTruncatedAndCipheredNAS(t *testing.T) {
	for _, packet := range [][]byte{nil, {0x7e}, {0x7e, 1}, {0x7e, 2, 0, 0, 0, 0, 0}} {
		_, err := DecodePlainNasNoIntegrityCheck(packet)
		require.Error(t, err)
	}
}
