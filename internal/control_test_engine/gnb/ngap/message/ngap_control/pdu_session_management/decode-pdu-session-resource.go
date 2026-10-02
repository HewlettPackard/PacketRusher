/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package pdu_session_management

import (
	"fmt"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	codec "my5G-RANTester/lib/ngap"
)

func GetGtpTeid(value *message.PDUSessionResourceSetupRequest) ([]byte, error) {
	if value == nil || value.PDUSessionResourceSetupListSUReq == nil {
		return nil, fmt.Errorf("missing PDU session setup list")
	}
	for _, item := range value.PDUSessionResourceSetupListSUReq.List {
		if item.PDUSessionResourceSetupRequestTransfer == nil {
			continue
		}
		transfer := &ie.PDUSessionResourceSetupRequestTransfer{}
		if err := codec.Unmarshal(*item.PDUSessionResourceSetupRequestTransfer, transfer); err != nil {
			return nil, err
		}
		if transfer.ProtocolIEs == nil {
			continue
		}
		for _, field := range transfer.ProtocolIEs.List {
			if field.ULNGUUPTNLInformation == nil {
				continue
			}
			tunnel, err := codec.Tunnel(field.ULNGUUPTNLInformation)
			if err != nil {
				return nil, err
			}
			return tunnel.GTPTEID.Value, nil
		}
	}
	return nil, fmt.Errorf("missing uplink GTP tunnel")
}
