package nas

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	NasForwarded "my5G-RANTester/internal/control_test_engine/ue/nas/message"
	"my5G-RANTester/internal/control_test_engine/ue/nas/message/nas_control/mm_5gs"
	"my5G-RANTester/lib/nas/nasMessage"
)

func InitRegistration(ue *context.UEContext) {

	// registration procedure started.
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(
		nasMessage.RegistrationType5GSInitialRegistration,
		nil,
		nil,
		ue)

	// send to GNB.
	NasForwarded.SendToGnb(ue, registrationRequest)

	// change the state of ue for deregistered
	ue.SetState(MM5G_DEREGISTERED)
}
