package ue

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas"
	"my5G-RANTester/internal/control_test_engine/ue/nas/service"
	"my5G-RANTester/lib/nas/security"
)

// init RegistrationUE(conn, imsi, int64(i), cfg, contextGnb, mcc, mnc)
// generate an ue data  and execute initial message registration.
func RegistrationUe(imsi string, conf config.Config, id uint8) {

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
		conf.Ue.Snssai.Sst,
		id)

	// starting communication with GNB.
	service.InitConn(ue)

	// TODO when unix sockets is closed.
	// defer conn.Close()

	// registration procedure started.
	nas.InitRegistration(ue)

	// listen GNB.
	service.UeListen(ue)
}
