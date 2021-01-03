package service

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap"
)

func InitConn(amf *context.GNBAmf, gnb *context.GNBContext) error {

	// check AMF IP and AMF port.
	sctpADDR := fmt.Sprintf("%s:%s", amf.GetAmfIp(), amf.GetAmfPort())

	addr, err := sctp.ResolveSCTPAddr("sctp", sctpADDR)
	if err != nil {
		return err
	}

	streams := amf.GetTNLAStreams()

	conn, err := sctp.DialSCTPExt(
		"sctp",
		nil,
		addr,
		sctp.InitMsg{NumOstreams: streams, MaxInstreams: streams})
	if err != nil {
		amf.SetSCTPConn(nil)
		return err
	}

	// set streams and other information about TNLA.

	// successful established SCTP
	amf.SetSCTPConn(conn)

	conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

	go GnBListen(amf, gnb)

	/*
		info := &sctp.SndRcvInfo{
			Stream: uint16(3),
			PPID:   ngapSctp.NGAP_PPID,
		}

		conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)

		conn.SCTPWrite([]byte("Oi"), info)

		time.Sleep(40*time.Second)
	*/

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

		n, err := conn.Read(buf[:])
		if err != nil {
			return
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NGAP message.
		go ngap.Dispatch(amf, gnb, forwardData)

	}

}
