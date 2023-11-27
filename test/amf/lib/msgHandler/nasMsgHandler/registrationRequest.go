package nasMsgHandler

import (
	"errors"
	"my5G-RANTester/test/amf/context"
	"strings"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/openapi/models"

	"github.com/free5gc/nas"
)

func RegistrationRequest(nasReq *nas.Message, amf *context.AMFContext, ue *context.UEContext) (err error) {
	regType := nasReq.RegistrationRequest.NgksiAndRegistrationType5GS.GetRegistrationType5GS()

	if regType != nasMessage.RegistrationType5GSInitialRegistration {
		return errors.New("[AMF][NAS] Received unsupported registration type")
	}

	gmm := nasReq.GmmMessage

	mobileId, mobileIdType, err := gmm.RegistrationRequest.MobileIdentity5GS.GetMobileIdentity()

	if mobileIdType != "SUCI" {
		return errors.New("[AMF][NAS] UE id uses IDType " + mobileIdType + " but is not yet supported by tests")
	}

	suci := strings.Split(mobileId, "-")
	sub, err := amf.FindSecurityContextByMsin(suci[len(suci)-1])

	if err != nil {
		return err
	}
	sub.SetSuci(mobileId)
	sub.SetSupi("imsi-" + suci[2] + suci[3] + suci[len(suci)-1])

	ue.SetSecurityContext(&sub)
	ue.SetSecurityCapability(gmm.RegistrationRequest.UESecurityCapability)
	//Todo: check if snssai is supported by amf, add possibility to request several NSSAI
	// snssai, err := nasConvert.RequestedNssaiToModels(gmm.RegistrationRequest.RequestedNSSAI)
	// ue.SetNssai(snssai[0])

	// NgKsi
	ngKsi := models.NgKsi{}
	switch nasReq.NgksiAndRegistrationType5GS.GetTSC() {
	case nasMessage.TypeOfSecurityContextFlagNative:
		ngKsi.Tsc = models.ScType_NATIVE
	default:
		return errors.New("[AMF] Unsupported KSI sc type")
	}
	ngKsi.Ksi = int32(nasReq.NgksiAndRegistrationType5GS.GetNasKeySetIdentifiler())
	if ngKsi.Tsc == models.ScType_NATIVE && ngKsi.Ksi != 7 {
	} else {
		ngKsi.Tsc = models.ScType_NATIVE
		ngKsi.Ksi = 0
	}

	ue.SetNgKsi(ngKsi)

	return nil
}
