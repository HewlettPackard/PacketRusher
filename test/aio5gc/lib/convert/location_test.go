package convert

import (
	"encoding/hex"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNGAPPLMNConversionsUseLiteralWireIdentity(t *testing.T) {
	for _, tc := range []struct{ mcc, mnc, wire string }{
		{"208", "93", "02f839"}, {"208", "010", "020801"},
		{"001", "001", "000110"}, {"001", "01", "00f110"},
		{"999", "070", "990907"}, {"999", "07", "99f970"},
	} {
		t.Run(tc.mcc+"/"+tc.mnc, func(t *testing.T) {
			wire, err := hex.DecodeString(tc.wire)
			require.NoError(t, err)
			literal := &ie.PLMNIdentity{Value: wire}
			model := models.PlmnId{Mcc: tc.mcc, Mnc: tc.mnc}
			require.Equal(t, wire, []byte(PLMNToNGAP(model).Value))
			require.Equal(t, model, PLMNToModels(literal))
			global, err := GlobalGNBToModels(&ie.GlobalGNBID{
				PLMNIdentity: literal,
				GNBID:        &ie.GNBID{Choice: &ie.GNBIDForGNBID{Value: aper.BitString{Bytes: []byte{0, 0, 1}, BitLength: 24}}},
			})
			require.NoError(t, err)
			require.Equal(t, &model, global.PlmnId)
			location := NRLocationToModels(&ie.UserLocationInformationNR{
				TAI:   &ie.TAI{PLMNIdentity: literal, TAC: &ie.TAC{Value: []byte{0, 0, 1}}},
				NRCGI: &ie.NRCGI{PLMNIdentity: literal, NRCellIdentity: &ie.NRCellIdentity{Value: aper.BitString{Bytes: []byte{0, 0, 0, 0, 0x10}, BitLength: 36}}},
			})
			require.Equal(t, &model, location.Tai.PlmnId)
			require.Equal(t, &model, location.Ncgi.PlmnId)
			served := models.PlmnIdNid{Mcc: tc.mcc, Mnc: tc.mnc}
			guami := GUAMIToNGAP(models.Guami{PlmnId: &served, AmfId: "000001"})
			require.Equal(t, wire, []byte(guami.PLMNIdentity.Value))
		})
	}
}

func TestPLMNConversionsRejectMalformedIdentity(t *testing.T) {
	require.Nil(t, PLMNToNGAP(models.PlmnId{Mcc: "001", Mnc: "0x"}))
	for _, value := range []*ie.PLMNIdentity{nil, {Value: []byte{0}}, {Value: []byte{0, 0xb1, 0x10}}} {
		require.Empty(t, PLMNToModels(value))
	}
}

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
