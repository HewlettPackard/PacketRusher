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

func PDUSessionResourceRelease(req *message.PDUSessionResourceReleaseResponse, fgc *context.Aio5gc) error {
	ue, err := resolveUE(fgc.GetAMFContext(), req.RANUENGAPID, req.AMFUENGAPID)
	if err != nil {
		return err
	}
	for _, item := range req.PDUSessionResourceReleasedListRelRes.List {
		if err := context.ConfirmPDUSessionRelease(ue, int32(item.PDUSessionID.Value)); err != nil {
			return err
		}
	}
	return nil
}
