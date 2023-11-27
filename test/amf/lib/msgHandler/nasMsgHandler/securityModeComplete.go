package nasMsgHandler

import (
	"errors"
	"fmt"
	"my5G-RANTester/test/amf/context"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
)

func SecurityModeComplete(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext) (err error) {
	securityModeComplete := nasReq.SecurityModeComplete
	if securityModeComplete.IMEISV != nil {
		if pei, err := nasConvert.PeiToStringWithError(securityModeComplete.IMEISV.Octet[:]); err != nil {
			return fmt.Errorf("[AMF][NAS] Decode PEI failed: %w", err)
		} else {
			ue.SetPei(pei)
		}
	}

	if securityModeComplete.NASMessageContainer == nil {
		return fmt.Errorf("[AMF][NAS] Empty NASMessageContainer in securityModeComplete message")
	}
	contents := securityModeComplete.NASMessageContainer.GetNASMessageContainerContents()
	m := nas.NewMessage()
	if err := m.GmmMessageDecode(&contents); err != nil {
		return err
	}

	switch m.GmmMessage.GmmHeader.GetMessageType() {
	case nas.MsgTypeRegistrationRequest:
		registrationRequest := m.GmmMessage.RegistrationRequest
		ue.SetSecurityCapability(registrationRequest.UESecurityCapability)
		ue.AllocateGuti(amf)
		ue.GetSecurityContext().UpdateSecurityContext()
		return nil
	default:
		return errors.New("nas message container Iei type error")
	}
}
