package service

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/state"
	"net"
)

func CloseConn(ue *context.UEContext) {
	conn := ue.GetUnixConn()
	conn.Close()
}

func InitConn(ue *context.UEContext) error {

	// initiated communication with GNB(unix sockets).
	conn, err := net.Dial("unix", "/tmp/gnb.sock")
	if err != nil {
		return fmt.Errorf("[UE] Error on Dial with server", err)
	}

	// store unix socket connection in the UE.
	ue.SetUnixConn(conn)

	// listen NAS.
	go UeListen(ue)

	return nil
}

// ue listen unix sockets.
func UeListen(ue *context.UEContext) {

	buf := make([]byte, 65535)
	conn := ue.GetUnixConn()

	/*
		defer func() {
			err := conn.Close()
			if err != nil {
				fmt.Printf("Error in closing unix sockets for %s ue\n", ue.GetSupi())
			}
		}()
	*/

	for {

		// read message.
		n, err := conn.Read(buf[:])
		if err != nil {
			break
		}

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// handling NAS message.
		go state.DispatchState(ue, forwardData)

	}
}
