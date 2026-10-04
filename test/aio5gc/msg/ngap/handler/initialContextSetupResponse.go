/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package handler

import (
	"github.com/free5gc/ngap/message"
	"my5G-RANTester/test/aio5gc/context"
)

func InitialContextSetupResponse(req *message.InitialContextSetupResponse, fgc *context.Aio5gc) error {
	_, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID)
	return err
}
