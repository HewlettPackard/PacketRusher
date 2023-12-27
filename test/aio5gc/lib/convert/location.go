/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package convert

import (
	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
)

func NRLocationToModels(location *ngapType.UserLocationInformationNR) *models.NrLocation {
	locationModel := models.NrLocation{}
	tai := ngapConvert.TaiToModels(location.TAI)
	plmn := ngapConvert.PlmnIdToModels(location.NRCGI.PLMNIdentity)
	ncgi := models.Ncgi{}
	ncgi.NrCellId = ngapConvert.BitStringToHex(&location.NRCGI.NRCellIdentity.Value)
	ncgi.PlmnId = &plmn
	locationModel.Tai = &tai
	locationModel.Ncgi = &ncgi
	return &locationModel
}
