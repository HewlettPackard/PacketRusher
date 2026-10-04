/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package convert

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	ngapCodec "my5G-RANTester/lib/ngap"
)

func GlobalGNBToModels(global *ie.GlobalGNBID) (models.GlobalRanNodeId, error) {
	if global == nil || global.GNBID == nil || global.PLMNIdentity == nil {
		return models.GlobalRanNodeId{}, fmt.Errorf("missing global gNB identity")
	}
	id, ok := global.GNBID.Choice.(*ie.GNBIDForGNBID)
	if !ok || id == nil || id.Value.BitLength < 22 || id.Value.BitLength > 32 || len(id.Value.Bytes) != int((id.Value.BitLength+7)/8) {
		return models.GlobalRanNodeId{}, fmt.Errorf("invalid gNB identity bit string")
	}
	mcc, mnc, err := ngapCodec.DecodePLMN(global.PLMNIdentity.Value)
	if err != nil {
		return models.GlobalRanNodeId{}, fmt.Errorf("invalid gNB PLMN: %w", err)
	}
	var value uint64
	for _, b := range id.Value.Bytes {
		value = value<<8 | uint64(b)
	}
	value >>= uint64(len(id.Value.Bytes)*8) - id.Value.BitLength
	return models.GlobalRanNodeId{
		PlmnId: &models.PlmnId{Mcc: mcc, Mnc: mnc},
		GNbId:  &models.GNbId{BitLength: int32(id.Value.BitLength), GNBValue: fmt.Sprintf("%0*x", (id.Value.BitLength+3)/4, value)},
	}, nil
}

func PLMNToModels(id *ie.PLMNIdentity) models.PlmnId {
	if id == nil {
		return models.PlmnId{}
	}
	mcc, mnc, _ := ngapCodec.DecodePLMN(id.Value)
	return models.PlmnId{Mcc: mcc, Mnc: mnc}
}
func PLMNToNGAP(id models.PlmnId) *ie.PLMNIdentity {
	b, err := ngapCodec.EncodePLMN(id.Mcc, id.Mnc)
	if err != nil {
		return nil
	}
	return &ie.PLMNIdentity{Value: b}
}
func SNSSAIToNGAP(id models.Snssai) *ie.SNSSAI {
	out := &ie.SNSSAI{SST: &ie.SST{Value: []byte{byte(id.Sst)}}}
	if id.Sd != "" {
		b, _ := hex.DecodeString(id.Sd)
		out.SD = &ie.SD{Value: b}
	}
	return out
}
func GUAMIToNGAP(id models.Guami) *ie.GUAMI {
	b, _ := hex.DecodeString(id.AmfId)
	return &ie.GUAMI{
		PLMNIdentity: PLMNToNGAP(models.PlmnId{Mcc: id.PlmnId.Mcc, Mnc: id.PlmnId.Mnc}),
		AMFRegionID:  &ie.AMFRegionID{Value: aper.BitString{Bytes: b[:1], BitLength: 8}},
		AMFSetID:     &ie.AMFSetID{Value: aper.BitString{Bytes: []byte{b[1], b[2] & 0xc0}, BitLength: 10}},
		AMFPointer:   &ie.AMFPointer{Value: aper.BitString{Bytes: []byte{b[2] << 2}, BitLength: 6}},
	}
}
func NRLocationToModels(location *ie.UserLocationInformationNR) *models.NrLocation {
	plmn := PLMNToModels(location.TAI.PLMNIdentity)
	cellplmn := PLMNToModels(location.NRCGI.PLMNIdentity)
	var cell uint64
	for _, b := range location.NRCGI.NRCellIdentity.Value.Bytes {
		cell = cell<<8 | uint64(b)
	}
	cell >>= uint64(len(location.NRCGI.NRCellIdentity.Value.Bytes)*8) - location.NRCGI.NRCellIdentity.Value.BitLength
	return &models.NrLocation{
		Tai:  &models.Tai{PlmnId: &plmn, Tac: hex.EncodeToString(location.TAI.TAC.Value)},
		Ncgi: &models.Ncgi{PlmnId: &cellplmn, NrCellId: fmt.Sprintf("%09x", cell)},
	}
}

func BitRate(value string) (int64, error) {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return 0, fmt.Errorf("invalid bit rate %q", value)
	}
	scales := map[string]int64{"bps": 1, "Kbps": 1000, "Mbps": 1000000, "Gbps": 1000000000, "Tbps": 1000000000000}
	scale, ok := scales[fields[1]]
	number, err := strconv.ParseInt(fields[0], 10, 64)
	if !ok || err != nil || number < 0 || number > 4000000000000/scale {
		return 0, fmt.Errorf("invalid bit rate %q", value)
	}
	return number * scale, nil
}
