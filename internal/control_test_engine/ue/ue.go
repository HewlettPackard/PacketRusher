package ue

import (
	"log"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/nas_control/mm_5gs"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/unix_sockets"
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/nas/security"
	"net"
)

// init RegistrationUE(conn, imsi, int64(i), cfg, contextGnb, mcc, mnc)
// generate an ue data  and execute initial message registration.
func registrationUe(imsi string, conf config.Config) {

	// new UE instance.
	ue := &context.UEContext{}

	// new UE context
	ue.NewRanUeContext(
		imsi,
		security.AlgCiphering128NEA0,
		security.AlgIntegrity128NIA2,
		conf.Ue.Key,
		conf.Ue.Opc,
		"c9e8763286b5b9ffbdf56e1297d0887b",
		conf.Ue.Amf,
		conf.Ue.Hplmn.Mcc,
		conf.Ue.Hplmn.Mnc,
		int32(conf.Ue.Snssai.Sd),
		conf.Ue.Snssai.Sst)

	// initiated communication with GNB(unix sockets).
	conn, err := net.Dial("unix", "/tmp/gnb.sock")
	if err != nil {
		log.Fatal("Dial error", err)
	}

	// stored unix socket connection in the UE.
	ue.SetUnixConn(conn)

	// registration procedure started.
	registrationRequest := mm_5gs.GetRegistrationRequestWith5GMM(
		nasMessage.RegistrationType5GSInitialRegistration,
		nil,
		nil,
		ue)

	// send to GNB.
	unix_sockets.SendToGnb(ue, registrationRequest)

	// listen GNB.
	unix_sockets.UeListen(ue)
}
