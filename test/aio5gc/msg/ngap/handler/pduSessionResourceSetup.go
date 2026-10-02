/** SPDX-License-Identifier: Apache-2.0 */
package handler

import (
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
)

func PDUSessionResourceSetup(req *message.PDUSessionResourceSetupResponse, fgc *context.Aio5gc) error {
	_, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID)
	return err
}
