package amf

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/context"
	amfTools "my5G-RANTester/test/amf/lib/tools"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestCreatePDUSession(t *testing.T) {

	controlIFConfig := config.ControlIF{
		Ip:   "192.168.11.13",
		Port: 9489,
	}
	dataIFConfig := config.DataIF{
		Ip:   "192.168.11.13",
		Port: 2154,
	}
	amfConfig := config.AMF{
		Ip:   "192.168.11.14",
		Port: 38414,
	}

	conf := amfTools.GenerateDefaultConf(controlIFConfig, dataIFConfig, amfConfig)

	// setup AMF with given handler
	amf := Amf{}
	ngapHandler := func(ngapMsg *ngapType.NGAPPDU, gnb context.GNBContext) (msg []byte, err error) {
		return amf.NgapDefaultHandler(ngapMsg, gnb)
	}
	nasHandler := func(nasPDU *ngapType.NASPDU, ue *context.UEContext) (uint8, error) {
		nasType, err := amf.NasDefaultHandler(nasPDU, ue)
		return nasType, err
	}
	err := amf.Init(conf, ngapHandler, nasHandler)
	if err != nil {
		log.Printf("[AMF] Error during amf initialization  %v", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)

	gnbCount := 1
	wg := sync.WaitGroup{}
	gnbs := tools.CreateGnbs(gnbCount, conf, &wg)

	time.Sleep(1 * time.Second)

	keys := make([]string, 0)
	for k := range gnbs {
		keys = append(keys, k)
	}

	ueCfg := conf

	securityContext := context.SecurityContext{}
	securityContext.SetMsin(ueCfg.Ue.Msin)
	securityContext.SetAuthSubscription(ueCfg.Ue.Key, ueCfg.Ue.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", conf.Ue.Amf, conf.Ue.Sqn)
	securityContext.SetAbba([]uint8{0x00, 0x00})
	amf.context.NewSecurityContext(securityContext)

	ueId := 1
	ueRx := make(chan procedures.UeTesterMessage)
	ueTx := ue.NewUE(ueCfg, uint8(ueId), ueRx, gnbs[keys[0]], &wg)

	// setup some PacketRusher UE
	ueRx <- procedures.UeTesterMessage{Type: procedures.Registration}
	log.Print(<-ueTx)
	log.Print(<-ueTx)
	log.Print(<-ueTx)
	log.Print(<-ueTx)

	// ueRx <- CreatePDUSession

	wg.Wait()
	// assert.True(t, true)
}
