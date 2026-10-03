// SPDX-License-Identifier: Apache-2.0
package types

import (
	"github.com/free5gc/openapi/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCloneTaiListPreservesAndIsolatesPrivateModels(t *testing.T) {
	source := []Tai{{Tac: "000001", plmnSnssaiList: []models.Nrf_NFMgmt_PlmnSnssai{{PlmnId: &models.PlmnId{Mcc: "208", Mnc: "93"}, SNssaiList: []models.ExtSnssai{{Sst: 1, Sd: "000001"}}}}}}
	clone := CloneTaiList(source)
	require.Equal(t, source, clone)
	clone[0].plmnSnssaiList[0].PlmnId.Mcc = "999"
	clone[0].plmnSnssaiList[0].SNssaiList[0].Sd = "ffffff"
	require.Equal(t, "208", source[0].plmnSnssaiList[0].PlmnId.Mcc)
	require.Equal(t, "000001", source[0].plmnSnssaiList[0].SNssaiList[0].Sd)
}
