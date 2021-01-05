package service

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"
)

func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {

	// check AMF IP and AMF port.
	remote := fmt.Sprintf("%s:%d", amf.GetAmfIp(), amf.GetAmfPort())
	local := fmt.Sprintf("%s:%d", gnb.GetGnbIp(), gnb.GetGnbPort())

	rem, err := sctp.ResolveSCTPAddr("sctp", remote)
	if err != nil {
		return err
	}
	loc, err := sctp.ResolveSCTPAddr("sctp", local)
	if err != nil {
		return err
	}

	// streams := amf.GetTNLAStreams()

	conn, err := sctp.DialSCTPExt(
		"sctp",
		loc,
		rem,
		sctp.InitMsg{NumOstreams: 1, MaxInstreams: 1})
	if err != nil {
		amf.SetSCTPConn(nil)
		return err
	}

	// set streams and other information about TNLA.

	// successful established SCTP.
	amf.SetSCTPConn(conn)

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

	go GnBListen(amf, gnb)

	return nil
}

func GnBListen(amf *context.GNBAmf, gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := amf.GetSCTPConn()

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error in closing SCTP association for %d AMF\n", amf.GetAmfId())
		}
	}()

	for {

		n, info, err := conn.SCTPRead(buf[:])
		if err != nil {
			return
		}

		fmt.Printf("Receive message in %d stream", info.Stream)

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go ngap.Dispatch(amf, gnb, forwardData)

	}

}
