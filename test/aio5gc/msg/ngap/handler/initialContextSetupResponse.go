/** SPDX-License-Identifier: Apache-2.0 */
package handler

import (
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
)

func InitialContextSetupResponse(req *message.InitialContextSetupResponse, fgc *context.Aio5gc, gnbs ...*context.GNBContext) error {
	_, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID, gnbs...)
	return err
}
