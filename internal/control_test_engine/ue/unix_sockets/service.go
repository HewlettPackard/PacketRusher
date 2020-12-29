package unix_sockets

import (
	"fmt"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas"
)

// ue listen unix sockets.
func UeListen(ue *context.UEContext) {

	buf := make([]byte, 65535)

	for {

		// read message.
		n, err := ue.UnixConn.Read(buf[:])
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

func SendToGnb(ue *context.UEContext, message []byte) {

	_, err := ue.UnixConn.Write(message)
	if err != nil {
		fmt.Println("Tratar o erro")
	}
}
