/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package builder

import (
	"fmt"
	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	"github.com/free5gc/util/fsm"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/test/aio5gc/context"
)

func UEContextReleaseCommand(ue *context.UEContext, present int, cause aper.Enumerated) ([]byte, error) {
	msg, err := buildUEContextReleaseCommand(ue, present, cause)
	if err != nil {
		return nil, err
	}
	b, err := msg.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = ue.GetUeFsm().SendEvent(ue.GetState(), context.Deregistration, fsm.ArgsType{"ue": ue}, log.NewEntry(log.StandardLogger()))
	return b, err
}
func buildUEContextReleaseCommand(ue *context.UEContext, present int, cause aper.Enumerated) (*message.UEContextReleaseCommand, error) {
	if present != 3 {
		return nil, fmt.Errorf("only NAS release causes supported")
	}
	return &message.UEContextReleaseCommand{UENGAPIDs: &ie.UENGAPIDs{Choice: &ie.UENGAPIDPair{AMFUENGAPID: &ie.AMFUENGAPID{Value: ue.GetAmfNgapId()}, RANUENGAPID: &ie.RANUENGAPID{Value: ue.GetRanNgapId()}}}, Cause: &ie.Cause{Choice: &ie.CauseNas{Value: cause}}}, nil
}
