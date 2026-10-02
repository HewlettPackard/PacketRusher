package convert

import (
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGlobalGNBIdentityRetainsBitLengthAndValue(t *testing.T) {
	for _, tt := range []struct {
		name  string
		bytes []byte
		bits  uint64
		want  string
	}{
		{"22-bit padded", []byte{0, 4, 8}, 22, "000102"},
		{"24-bit aligned", []byte{0, 1, 2}, 24, "000102"},
		{"32-bit aligned", []byte{1, 0x23, 0x45, 0x67}, 32, "01234567"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			value, err := GlobalGNBToModels(&ie.GlobalGNBID{PLMNIdentity: PLMNToNGAP(models.PlmnId{Mcc: "208", Mnc: "93"}), GNBID: &ie.GNBID{Choice: &ie.GNBIDForGNBID{Value: aper.BitString{Bytes: tt.bytes, BitLength: tt.bits}}}})
			require.NoError(t, err)
			require.Equal(t, &models.PlmnId{Mcc: "208", Mnc: "93"}, value.PlmnId)
			require.Equal(t, int32(tt.bits), value.GNbId.BitLength)
			require.Equal(t, tt.want, value.GNbId.GNBValue)
		})
	}
}

func TestGlobalGNBIdentityRejectsInvalidInput(t *testing.T) {
	for _, id := range []*ie.GlobalGNBID{
		nil, {},
		{PLMNIdentity: &ie.PLMNIdentity{Value: []byte{0}}, GNBID: &ie.GNBID{Choice: &ie.GNBIDForGNBID{Value: aper.BitString{Bytes: []byte{0, 0, 0}, BitLength: 24}}}},
		{PLMNIdentity: PLMNToNGAP(models.PlmnId{Mcc: "208", Mnc: "93"}), GNBID: &ie.GNBID{Choice: &ie.GNBIDForGNBID{Value: aper.BitString{Bytes: []byte{0}, BitLength: 24}}}},
	} {
		_, err := GlobalGNBToModels(id)
		require.Error(t, err)
	}
}
