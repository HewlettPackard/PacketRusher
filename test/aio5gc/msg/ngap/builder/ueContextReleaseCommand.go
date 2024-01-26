/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"fmt"
	"my5G-RANTester/test/aio5gc/context"
	"my5G-RANTester/test/aio5gc/state"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap"
	"github.com/free5gc/util/fsm"

	"github.com/free5gc/ngap/ngapType"
	log "github.com/sirupsen/logrus"
)

func UEContextReleaseCommand(ue *context.UEContext, causePresent int, cause aper.Enumerated) ([]byte, error) {

	msg, err := buildUEContextReleaseCommand(ue, causePresent, cause)
	if err != nil {
		return nil, err
	}
	pdu, err := ngap.Encoder(msg)
	if err != nil {
		return nil, err
	}
	err = state.GetUeFsm().SendEvent(ue.GetState(), state.Deregistration, fsm.ArgsType{"ue": ue}, log.NewEntry(log.StandardLogger()))
	if err != nil {
		return nil, err
	}
	return pdu, nil
}

func buildUEContextReleaseCommand(ue *context.UEContext, causePresent int, cause aper.Enumerated) (ngapType.NGAPPDU, error) {
	var pdu ngapType.NGAPPDU

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeUEContextRelease
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentUEContextReleaseCommand
	initiatingMessage.Value.UEContextReleaseCommand = new(ngapType.UEContextReleaseCommand)

	ueContextReleaseCommand := initiatingMessage.Value.UEContextReleaseCommand
	ueContextReleaseCommandIEs := &ueContextReleaseCommand.ProtocolIEs

	// UE NGAP IDs
	ie := ngapType.UEContextReleaseCommandIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUENGAPIDs
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.UEContextReleaseCommandIEsPresentUENGAPIDs
	ie.Value.UENGAPIDs = new(ngapType.UENGAPIDs)

	ueNGAPIDs := ie.Value.UENGAPIDs

	ueNGAPIDs.Present = ngapType.UENGAPIDsPresentUENGAPIDPair
	ueNGAPIDs.UENGAPIDPair = new(ngapType.UENGAPIDPair)

	ueNGAPIDs.UENGAPIDPair.AMFUENGAPID.Value = ue.GetAmfNgapId()
	ueNGAPIDs.UENGAPIDPair.RANUENGAPID.Value = ue.GetRanNgapId()

	ueContextReleaseCommandIEs.List = append(ueContextReleaseCommandIEs.List, ie)

	// Cause
	ie = ngapType.UEContextReleaseCommandIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDCause
	ie.Criticality.Value = ngapType.CriticalityPresentIgnore
	ie.Value.Present = ngapType.UEContextReleaseCommandIEsPresentCause
	ngapCause := ngapType.Cause{
		Present: causePresent,
	}
	switch causePresent {
	case ngapType.CausePresentNas:
		ngapCause.Nas = new(ngapType.CauseNas)
		ngapCause.Nas.Value = cause
	default:
		return ngapType.NGAPPDU{}, fmt.Errorf("[5GC] Deregistration Cause Present is Unknown")
	}
	ie.Value.Cause = &ngapCause

	ueContextReleaseCommandIEs.List = append(ueContextReleaseCommandIEs.List, ie)

	return pdu, nil
}
