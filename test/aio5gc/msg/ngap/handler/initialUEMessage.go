/** SPDX-License-Identifier: Apache-2.0 */
package handler

import (
	"fmt"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
	"my5G-RANTester/test/aio5gc/msg/nas"
)

func InitialUEMessage(req *message.InitialUEMessage, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	location, ok := req.UserLocationInformation.Choice.(*ie.UserLocationInformationNR)
	if !ok {
		return fmt.Errorf("mock core requires NR location")
	}
	ue := fgc.GetAMFContext().NewUE(req.RANUENGAPID.Value)
	model := convert.NRLocationToModels(location)
	model.GlobalGnbId = gnb.GetGlobalRanNodeID()
	ue.SetUserLocationInfo(model)
	nas.Dispatch(req.NASPDU, ue, fgc, gnb)
	return nil
}
