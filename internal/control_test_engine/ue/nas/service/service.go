package service

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas"
	"net"
)

func closeConn(ue *context.UEContext) {
	conn := ue.GetUnixConn()
	conn.Close()
}

func InitConn(ue *context.UEContext) {

	// initiated communication with GNB(unix sockets).
	conn, err := net.Dial("unix", "/tmp/gnb.sock")
	if err != nil {
		//log.Fatal("Dial error", err)
	}

	// store unix socket connection in the UE.
	ue.SetUnixConn(conn)

	// change the state of ue for deregistered
	ue.SetState(0x01)
}

// ue listen unix sockets.
func UeListen(ue *context.UEContext) {

	buf := make([]byte, 65535)
	conn := ue.GetUnixConn()

	for {

		// read message.
		n, err := conn.Read(buf[:])
		if err != nil {
			return
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NAS message.
		go nas.Dispatch(ue, forwardData)

		//fmt.Println( fmt.Sprintf("Client : %s received: %s", s, string(buf[0:n]) ) )
	}
}
