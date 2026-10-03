/** SPDX-License-Identifier: Apache-2.0 */
package handler

import (
	"fmt"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/lib/convert"
	"my5G-RANTester/test/aio5gc/lib/types"
	"my5G-RANTester/test/aio5gc/msg"
)

func NGSetupRequest(req *message.NGSetupRequest, gnb *context.GNBContext, fgc *context.Aio5gc) error {
	global, ok := req.GlobalRANNodeID.Choice.(*ie.GlobalGNBID)
	if !ok {
		return fmt.Errorf("mock core requires a gNB identity")
	}
	identity, err := convert.GlobalGNBToModels(global)
	if err != nil {
		return err
	}
	gnb.SetGlobalRanNodeID(identity)
	if req.RANNodeName != nil {
		gnb.SetRanNodename(string(req.RANNodeName.Value))
	}
	gnb.SetSuportedTAList(types.TaiListToModels(*req.SupportedTAList))
	if req.DefaultPagingDRX != nil {
		gnb.SetDefautlPagingDRX(*req.DefaultPagingDRX)
	}
	return msg.SendNGSetupResponse(gnb, fgc.GetAMFContext())
}
