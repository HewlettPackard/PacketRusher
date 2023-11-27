package nasMsgHandler

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"my5G-RANTester/test/amf/context"
	"my5G-RANTester/test/amf/lib/tools"

	"github.com/free5gc/nas"
	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/openapi/models"
	"github.com/mitchellh/mapstructure"
)

func AuthenticationRequest(ue *context.UEContext) (msg []byte, err error) {
	plmn := *ue.GetUserLocationInfo().Tai.PlmnId
	var servingNetworkName string
	if len(plmn.Mnc) == 2 {
		servingNetworkName = "5G:mnc0" + plmn.Mnc + ".mcc" + plmn.Mcc + ".3gppnetwork.org"
	} else {
		servingNetworkName = "5G:mnc" + plmn.Mnc + ".mcc" + plmn.Mcc + ".3gppnetwork.org"
	}

	authResult, _ := tools.AuthProcedure(ue.GetSecurityContext().GetAuthSubscription(), servingNetworkName)
	ue.GetSecurityContext().SetXresStar(authResult.AuthenticationVector.XresStar)

	authCtx, kseaf, err := tools.DeriveHXRES(authResult, servingNetworkName)
	if err != nil {
		return msg, err
	}
	ue.GetSecurityContext().SetKseaf(kseaf)

	return buildAuthenticationRequest(*ue, authCtx)
}

func buildAuthenticationRequest(ue context.UEContext, authCtx models.UeAuthenticationCtx) (msg []byte, err error) {
	m := nas.NewMessage()
	m.GmmMessage = nas.NewGmmMessage()
	m.GmmHeader.SetMessageType(nas.MsgTypeAuthenticationRequest)

	authenticationRequest := nasMessage.NewAuthenticationRequest(0)
	authenticationRequest.SetExtendedProtocolDiscriminator(nasMessage.Epd5GSMobilityManagementMessage)
	authenticationRequest.SpareHalfOctetAndSecurityHeaderType.SetSecurityHeaderType(nas.SecurityHeaderTypePlainNas)
	authenticationRequest.SpareHalfOctetAndSecurityHeaderType.SetSpareHalfOctet(0)
	authenticationRequest.AuthenticationRequestMessageIdentity.SetMessageType(nas.MsgTypeAuthenticationRequest)
	authenticationRequest.SpareHalfOctetAndNgksi = nasConvert.SpareHalfOctetAndNgksiToNas(ue.GetNgKsi())

	// Dummy abba, not supported by PR for now
	abba := []uint8{0x00, 0x00}
	authenticationRequest.ABBA.SetLen(uint8(len(abba)))
	authenticationRequest.ABBA.SetABBAContents(abba)

	switch authCtx.AuthType {
	case models.AuthType__5_G_AKA:
		var tmpArray [16]byte
		var av5gAka models.Av5gAka

		if err := mapstructure.Decode(authCtx.Var5gAuthData, &av5gAka); err != nil {
			return nil, errors.New("[AMF]Var5gAuthData Convert Type Error")
		}

		rand, err := hex.DecodeString(av5gAka.Rand)
		if err != nil {
			return nil, err
		}
		authenticationRequest.AuthenticationParameterRAND = nasType.
			NewAuthenticationParameterRAND(nasMessage.AuthenticationRequestAuthenticationParameterRANDType)
		copy(tmpArray[:], rand[0:16])
		authenticationRequest.AuthenticationParameterRAND.SetRANDValue(tmpArray)

		autn, err := hex.DecodeString(av5gAka.Autn)
		if err != nil {
			return nil, err
		}
		authenticationRequest.AuthenticationParameterAUTN = nasType.
			NewAuthenticationParameterAUTN(nasMessage.AuthenticationRequestAuthenticationParameterAUTNType)
		authenticationRequest.AuthenticationParameterAUTN.SetLen(uint8(len(autn)))
		copy(tmpArray[:], autn[0:16])
		authenticationRequest.AuthenticationParameterAUTN.SetAUTN(tmpArray)
	default:
		return msg, errors.New("[AMF] AuthenticationRequest unsupported AuthType")
	}

	m.GmmMessage.AuthenticationRequest = authenticationRequest

	data := new(bytes.Buffer)
	err = m.GmmMessageEncode(data)
	if err != nil {
		fmt.Println(err.Error())
	}

	msg = data.Bytes()
	return

}
